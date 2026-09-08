# Standard shipping

Make `python3 quote.py SUBTOTAL standard` charge 800 cents below 50.00 and
zero at or above 50.00. Print JSON with `subtotal_cents`, `shipping_cents`,
`total_cents` equal to their sum, and `mode: "standard"`.

Demonstrate 49.99, 50.00, and 50.01 through the CLI. Expedited shipping is
the next sprint; leave it unchanged in this sprint.
