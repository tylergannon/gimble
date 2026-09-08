# Expedited shipping

Make `python3 quote.py SUBTOTAL expedited` charge 1200 cents at every
subtotal. Print the same JSON shape as standard shipping, with
`mode: "expedited"`. Preserve standard shipping's inclusive 50.00 threshold.

Demonstrate both modes at 49.99, 50.00, and 50.01 through the CLI.
