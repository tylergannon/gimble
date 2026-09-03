// Package checklist parses and updates checklist files: markdown documents
// whose YAML frontmatter is a ledger of items the loop node works through.
//
// The frontmatter is the only part the engine reads. The markdown body is
// prose for agents and people and is preserved byte for byte on rewrite.
package checklist

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Item is one checklist entry. Done is engine-owned.
type Item struct {
	Name      string
	Check     string
	Command   string
	Infer     *Infer
	Doc       string
	Checklist string
	Done      bool
	// DonePresent distinguishes an engine-authored done field from an absent
	// field. Planning artifact validation uses it to reject agent-authored
	// done: false as well as done: true.
	DonePresent bool
}

// Infer describes evidence files and the question a judge answers about them.
type Infer struct {
	Files  []string
	Prompt string
}

// Checklist is a parsed checklist file.
type Checklist struct {
	Path  string
	Items []Item
	Body  string // the markdown after the closing ---, verbatim
}

// Load parses the file at path.
func Load(path string) (*Checklist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("checklist %s: %w", path, err)
	}
	return Parse(path, data)
}

// Parse parses checklist bytes (path is used only in error messages).
func Parse(path string, data []byte) (*Checklist, error) {
	list, _, err := parse(path, data)
	return list, err
}

// Open returns the first item whose Done is false, with its zero-based index,
// or ok=false when every item is done.
func (c *Checklist) Open() (item Item, index int, ok bool) {
	for i, it := range c.Items {
		if !it.Done {
			return it, i, true
		}
	}
	return Item{}, -1, false
}

// Find returns the item with the given name and its index.
func (c *Checklist) Find(name string) (item Item, index int, ok bool) {
	for i, it := range c.Items {
		if it.Name == name {
			return it, i, true
		}
	}
	return Item{}, -1, false
}

// MarkDone rewrites the file at path so that the item named name has
// done: true. The markdown body is preserved byte for byte. The frontmatter
// is re-encoded from the parsed YAML node tree so key order and comments
// survive as far as the library allows. Returns an error if no item has that
// name.
func MarkDone(path, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("checklist %s: %w", path, err)
	}
	list, doc, err := parse(path, data)
	if err != nil {
		return err
	}
	if _, _, ok := list.Find(name); !ok {
		return fmt.Errorf("checklist %s: no item named %q", path, name)
	}

	items := mappingValue(doc.root, "items")
	var target *yaml.Node
	for _, entry := range items.Content {
		if key := mappingValue(entry, "name"); key != nil && key.Value == name {
			target = entry
			break
		}
	}
	if target == nil {
		return fmt.Errorf("checklist %s: no item named %q", path, name)
	}
	setDone(target)

	var buf bytes.Buffer
	buf.WriteString("---\n")
	encoded, err := yaml.Dump(doc.node, dumpOptions...)
	if err != nil {
		return fmt.Errorf("checklist %s: encode frontmatter: %w", path, err)
	}
	buf.Write(encoded)
	if len(encoded) > 0 && encoded[len(encoded)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString("---")
	buf.WriteString(doc.body)

	// Prove the rewrite is valid before it touches the file.
	rewritten, err := Parse(path, buf.Bytes())
	if err != nil {
		return fmt.Errorf("checklist %s: rewritten frontmatter does not parse: %w", path, err)
	}
	if rewritten.Body != doc.body {
		return fmt.Errorf("checklist %s: rewrite would alter the markdown body", path)
	}
	if marked, _, ok := rewritten.Find(name); !ok || !marked.Done {
		return fmt.Errorf("checklist %s: rewrite did not mark %q done", path, name)
	}
	return writeAtomic(path, buf.Bytes())
}

// Render returns the item as a compact YAML block for prompt injection,
// omitting absent fields and never printing done.
func (i Item) Render() string {
	type inferView struct {
		Prompt string   `yaml:"prompt"`
		Files  []string `yaml:"files"`
	}
	type itemView struct {
		Name      string     `yaml:"name"`
		Check     string     `yaml:"check"`
		Command   string     `yaml:"command,omitempty"`
		Infer     *inferView `yaml:"infer,omitempty"`
		Doc       string     `yaml:"doc,omitempty"`
		Checklist string     `yaml:"checklist,omitempty"`
	}
	view := itemView{
		Name:      i.Name,
		Check:     i.Check,
		Command:   i.Command,
		Doc:       i.Doc,
		Checklist: i.Checklist,
	}
	if i.Infer != nil {
		view.Infer = &inferView{Prompt: i.Infer.Prompt, Files: i.Infer.Files}
	}
	out, err := yaml.Dump(view, dumpOptions...)
	if err != nil {
		// A struct of strings and string slices always encodes; keep the
		// signature simple and fall back to the bare name.
		return "name: " + i.Name
	}
	return strings.TrimRight(string(out), "\n")
}

// dumpOptions shape every YAML this package writes: two-space indent,
// sequences indented under their key the way people hand-write them, and no
// line folding so long prose stays on one line.
var dumpOptions = []yaml.Option{
	yaml.WithV4Defaults(),
	yaml.WithIndent(2),
	yaml.WithCompactSeqIndent(false),
	yaml.WithLineWidth(-1),
}

// document is the parsed frontmatter tree plus the untouched body.
type document struct {
	node *yaml.Node // the DocumentNode
	root *yaml.Node // the top-level mapping
	body string
}

var itemKeys = map[string]bool{
	"name": true, "check": true, "command": true, "infer": true,
	"doc": true, "checklist": true, "done": true,
}

var inferKeys = map[string]bool{"files": true, "prompt": true}

func parse(path string, data []byte) (*Checklist, *document, error) {
	front, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, nil, fmt.Errorf("checklist %s: %w", path, err)
	}

	var node yaml.Node
	if err := yaml.Load(front, &node, yaml.WithV4Defaults()); err != nil {
		return nil, nil, fmt.Errorf("checklist %s: frontmatter: %w", path, err)
	}
	if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
		return nil, nil, fmt.Errorf("checklist %s: frontmatter must be a single YAML mapping", path)
	}
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("checklist %s: frontmatter must be a YAML mapping", path)
	}

	items := mappingValue(root, "items")
	if items == nil {
		return nil, nil, fmt.Errorf("checklist %s: frontmatter has no items key", path)
	}
	if items.Kind != yaml.SequenceNode {
		return nil, nil, fmt.Errorf("checklist %s: items must be a list (line %d)", path, items.Line)
	}

	list := &Checklist{Path: path, Body: body, Items: make([]Item, 0, len(items.Content))}
	seen := make(map[string]int, len(items.Content))
	for i, entry := range items.Content {
		item, err := parseItem(entry)
		if err != nil {
			return nil, nil, fmt.Errorf("checklist %s: item %d: %w", path, i+1, err)
		}
		if prev, dup := seen[item.Name]; dup {
			return nil, nil, fmt.Errorf("checklist %s: item %d: duplicate name %q (also item %d)", path, i+1, item.Name, prev+1)
		}
		seen[item.Name] = i
		list.Items = append(list.Items, item)
	}
	return list, &document{node: &node, root: root, body: body}, nil
}

func parseItem(entry *yaml.Node) (Item, error) {
	if entry.Kind != yaml.MappingNode {
		return Item{}, fmt.Errorf("must be a mapping (line %d)", entry.Line)
	}
	var item Item
	seen := map[string]bool{}
	for k := 0; k+1 < len(entry.Content); k += 2 {
		keyNode, valNode := entry.Content[k], entry.Content[k+1]
		key := keyNode.Value
		if !itemKeys[key] {
			return Item{}, fmt.Errorf("unknown key %q (line %d)", key, keyNode.Line)
		}
		if seen[key] {
			return Item{}, fmt.Errorf("duplicate key %q (line %d)", key, keyNode.Line)
		}
		seen[key] = true

		var err error
		switch key {
		case "name":
			item.Name, err = scalar(valNode, key)
		case "check":
			item.Check, err = scalar(valNode, key)
		case "command":
			item.Command, err = scalar(valNode, key)
		case "doc":
			item.Doc, err = scalar(valNode, key)
		case "checklist":
			item.Checklist, err = scalar(valNode, key)
		case "done":
			item.DonePresent = true
			item.Done, err = boolean(valNode, key)
		case "infer":
			item.Infer, err = parseInfer(valNode)
		}
		if err != nil {
			return Item{}, err
		}
	}
	if strings.TrimSpace(item.Name) == "" {
		return Item{}, fmt.Errorf("name is required (line %d)", entry.Line)
	}
	if strings.TrimSpace(item.Check) == "" {
		return Item{}, fmt.Errorf("item %q: check is required", item.Name)
	}
	if item.Infer != nil {
		if len(item.Infer.Files) == 0 {
			return Item{}, fmt.Errorf("item %q: infer.files must list at least one path", item.Name)
		}
		if strings.TrimSpace(item.Infer.Prompt) == "" {
			return Item{}, fmt.Errorf("item %q: infer.prompt is required", item.Name)
		}
	}
	return item, nil
}

func parseInfer(node *yaml.Node) (*Infer, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("infer must be a mapping (line %d)", node.Line)
	}
	infer := &Infer{}
	seen := map[string]bool{}
	for k := 0; k+1 < len(node.Content); k += 2 {
		keyNode, valNode := node.Content[k], node.Content[k+1]
		key := keyNode.Value
		if !inferKeys[key] {
			return nil, fmt.Errorf("infer: unknown key %q (line %d)", key, keyNode.Line)
		}
		if seen[key] {
			return nil, fmt.Errorf("infer: duplicate key %q (line %d)", key, keyNode.Line)
		}
		seen[key] = true
		switch key {
		case "prompt":
			prompt, err := scalar(valNode, "infer.prompt")
			if err != nil {
				return nil, err
			}
			infer.Prompt = prompt
		case "files":
			files, err := stringList(valNode, "infer.files")
			if err != nil {
				return nil, err
			}
			infer.Files = files
		}
	}
	return infer, nil
}

// scalar returns the string value of a scalar node; null yields "".
func scalar(node *yaml.Node, key string) (string, error) {
	if node.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("%s must be a string (line %d)", key, node.Line)
	}
	if node.ShortTag() == "!!null" {
		return "", nil
	}
	return node.Value, nil
}

func boolean(node *yaml.Node, key string) (bool, error) {
	if node.Kind != yaml.ScalarNode || node.ShortTag() != "!!bool" {
		return false, fmt.Errorf("%s must be true or false (line %d)", key, node.Line)
	}
	return strings.EqualFold(node.Value, "true"), nil
}

// stringList accepts one string or a list of strings; blanks are rejected.
func stringList(node *yaml.Node, key string) ([]string, error) {
	var raw []*yaml.Node
	switch node.Kind {
	case yaml.ScalarNode:
		if node.ShortTag() == "!!null" {
			return nil, nil
		}
		raw = []*yaml.Node{node}
	case yaml.SequenceNode:
		raw = node.Content
	default:
		return nil, fmt.Errorf("%s must be a string or a list of strings (line %d)", key, node.Line)
	}
	out := make([]string, 0, len(raw))
	for _, n := range raw {
		if n.Kind != yaml.ScalarNode || n.ShortTag() == "!!null" {
			return nil, fmt.Errorf("%s must be a string or a list of strings (line %d)", key, n.Line)
		}
		if strings.TrimSpace(n.Value) == "" {
			return nil, fmt.Errorf("%s contains a blank entry (line %d)", key, n.Line)
		}
		out = append(out, n.Value)
	}
	return out, nil
}

// mappingValue returns the value node for key in a mapping, or nil.
func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for k := 0; k+1 < len(mapping.Content); k += 2 {
		if mapping.Content[k].Value == key {
			return mapping.Content[k+1]
		}
	}
	return nil
}

// setDone sets done: true on an item mapping, appending the key if absent.
func setDone(item *yaml.Node) {
	if val := mappingValue(item, "done"); val != nil {
		val.Kind = yaml.ScalarNode
		val.Tag = "!!bool"
		val.Value = "true"
		val.Style = 0
		val.Content = nil
		return
	}
	item.Content = append(item.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "done"},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"},
	)
}

// splitFrontmatter separates the YAML between the opening and closing ---
// lines from the body. The body begins immediately after the closing ---
// characters, so it carries that line's newline verbatim.
func splitFrontmatter(data []byte) (front []byte, body string, err error) {
	start, ok := afterMarkerLine(data, 0)
	if !ok {
		return nil, "", errors.New("missing frontmatter: file must start with a --- line")
	}
	pos := start
	for pos <= len(data) {
		if isMarkerLine(data, pos) {
			return data[start:pos], string(data[pos+3:]), nil
		}
		next := bytes.IndexByte(data[pos:], '\n')
		if next < 0 {
			break
		}
		pos += next + 1
	}
	return nil, "", errors.New("missing frontmatter: no closing --- line")
}

// isMarkerLine reports whether the line starting at pos is exactly ---.
func isMarkerLine(data []byte, pos int) bool {
	if !bytes.HasPrefix(data[pos:], []byte("---")) {
		return false
	}
	rest := data[pos+3:]
	return len(rest) == 0 || rest[0] == '\n' || bytes.HasPrefix(rest, []byte("\r\n"))
}

// afterMarkerLine returns the offset just past the marker line at pos.
func afterMarkerLine(data []byte, pos int) (int, bool) {
	if !isMarkerLine(data, pos) {
		return 0, false
	}
	end := pos + 3
	if end < len(data) && data[end] == '\r' {
		end++
	}
	if end < len(data) && data[end] == '\n' {
		end++
	}
	return end, true
}

// writeAtomic replaces the file at path by renaming a temp file over it. A
// symlink is followed so the target is rewritten and the link survives; the
// existing file's permission bits are preserved (0644 only when the file did
// not exist).
func writeAtomic(path string, data []byte) error {
	target := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		target = resolved
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checklist %s: %w", path, err)
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(target); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checklist %s: %w", path, err)
	}
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(target)+".*.tmp")
	if err != nil {
		return fmt.Errorf("checklist %s: %w", path, err)
	}
	tmpPath := tmp.Name()
	fail := func(err error) error {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("checklist %s: %w", path, err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fail(err)
	}
	if err := tmp.Chmod(mode); err != nil {
		return fail(err)
	}
	if err := tmp.Sync(); err != nil {
		return fail(err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("checklist %s: %w", path, err)
	}
	if err := os.Rename(tmpPath, target); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("checklist %s: %w", path, err)
	}
	return nil
}
