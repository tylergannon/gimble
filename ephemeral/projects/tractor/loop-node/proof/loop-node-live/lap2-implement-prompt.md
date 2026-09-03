<tractor loop="items" checklist="ephemeral/projects/demo/checklist.md" item="2/3" lap="1">
name: line count
check: count.txt contains the number of lines in greet.sh, as a bare integer
command: test "$(cat count.txt)" = "$(wc -l < greet.sh | tr -d ' ')"
</tractor>

Implement the current checklist item so that its check holds.