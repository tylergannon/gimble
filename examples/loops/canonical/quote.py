#!/usr/bin/env python3
"""Starting application: standard has a boundary bug; expedited is missing."""
import json
import sys
from decimal import Decimal

subtotal = int(Decimal(sys.argv[1]) * 100)
mode = sys.argv[2]
if mode != "standard":
    raise SystemExit("unsupported shipping mode")
shipping = 0 if subtotal > 5000 else 800
print(json.dumps({"subtotal_cents": subtotal, "shipping_cents": shipping,
                  "total_cents": subtotal + shipping, "mode": mode}))
