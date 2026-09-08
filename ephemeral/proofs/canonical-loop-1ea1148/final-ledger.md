---
items:
  - name: Standard shipping
    check: Standard shipping costs 800 cents below 50.00 and zero at or above 50.00; the JSON subtotal, shipping, total and mode are correct.
    command: python3 check.py standard
    doc: docs/sprints/standard.md
    infer:
      files: evidence-standard.json
      prompt: Do the recorded CLI invocations demonstrate the standard shipping promise at 49.99, 50.00 and 50.01? Judge the actual output and exit codes.
    done: true
  - name: Expedited shipping
    check: Expedited shipping costs 1200 cents at 49.99, 50.00 and 50.01; the JSON subtotal, shipping, total and mode are correct.
    command: python3 check.py expedited
    doc: docs/sprints/expedited.md
    infer:
      files: evidence-expedited.json
      prompt: Do the recorded CLI invocations demonstrate the expedited shipping promise at 49.99, 50.00 and 50.01? Judge the actual output and exit codes.
    done: true
---

# Definition of done

The shipping quote CLI works in both standard and expedited mode at 49.99,
50.00 and 50.01. Standard shipping is 800 cents below 50.00 and free at or
above it. Expedited shipping is always 1200 cents. Every successful invocation
prints JSON with the correct subtotal_cents, shipping_cents, total_cents and
mode. Both ledger items must be demonstrated before this goal is done.

These two sprints are the complete scope. Preserve their acceptance criteria.
