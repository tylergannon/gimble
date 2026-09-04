#!/usr/bin/env python3
"""Deliberately faulty starting application for the negative live proof."""
import json
import sys
from decimal import Decimal

subtotal = int(Decimal(sys.argv[1]) * 100)
mode = sys.argv[2]
if mode != "standard":
    raise SystemExit("unsupported shipping mode")
# Fault injection: exactly $50 incorrectly incurs shipping.
shipping = 0 if subtotal > 5000 else 800
print(json.dumps({"subtotal_cents": subtotal, "shipping_cents": shipping,
                  "total_cents": subtotal + shipping, "mode": mode}))
