#!/usr/bin/env python3
"""Shipping quote CLI: standard is free at or above 50.00, expedited is flat."""
import json
import sys
from decimal import Decimal

subtotal = int(Decimal(sys.argv[1]) * 100)
mode = sys.argv[2]
if mode == "standard":
    shipping = 0 if subtotal >= 5000 else 800
elif mode == "expedited":
    shipping = 1200
else:
    raise SystemExit("unsupported shipping mode")
print(json.dumps({"subtotal_cents": subtotal, "shipping_cents": shipping,
                  "total_cents": subtotal + shipping, "mode": mode}))
