---
items:
  - name: Standard shipping boundary
    check: >-
      Running quote.py with mode standard charges 800 cents for subtotal 49.99,
      zero shipping for subtotal 50.00, and zero shipping for subtotal 50.01.
      Each successful JSON result reports subtotal_cents, shipping_cents,
      total_cents equal to their sum, and mode standard.
    command: python3 capture.py
    infer:
      files: evidence.json
      prompt: >-
        Inspect the actual standard-mode invocation results. The inclusive
        50.00 threshold matters. Missing, failed, or incorrect standard-mode
        results fail this item. Expedited mode belongs to the overall goal,
        not this item. You may independently run quote.py to check the evidence.
---

# Definition of done

The shipping quote CLI must work in both standard and expedited mode:

- `python3 quote.py SUBTOTAL standard` charges 800 cents when SUBTOTAL is less
  than 50.00, and zero when it is at least 50.00.
- `python3 quote.py SUBTOTAL expedited` charges 1200 cents at every subtotal.
- On each successful invocation it prints a JSON object with `subtotal_cents`,
  `shipping_cents`, `total_cents` (their sum), and `mode`.
- Operate both modes at 49.99, 50.00, and 50.01 before deciding the goal is met.

The initial ledger intentionally contains only the standard-shipping item.
Passing that item is not completion: if expedited mode is missing, leave a
concrete open item for it, with a runnable verification command, and return
not_done. Edit only open ledger items and preserve this definition of done.
