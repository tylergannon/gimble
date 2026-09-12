# Third-party notice

`internal/sessionstate` is a hand-written Go port of session-projection code
from OpenCode, at commit `c55ee2a8152603f04a409163bd3edf79c425fbd7`.

Ported sources and their git blob hashes at that commit:

| Source | Blob |
| --- | --- |
| `packages/client/src/solid/data.ts` | `83c6eeb32ecccb1e4e9fb589647ea8e33f96b71e` |
| `packages/schema/src/session-event.ts` | `6e43890ff8ae67d88f4990d710875f0d67b8db65` |
| `packages/schema/src/session-message.ts` | `7e8ce0482ae9db77bb1ddc8e9bb3fb0b4c41dcdf` |
| `packages/schema/src/session-inbox.ts` | `baadcc947b18c267e10fee76cb648f140ccdffe6` |
| `packages/schema/src/session.ts` | read at the same commit |
| `packages/schema/src/event.ts` | read at the same commit |
| `LICENSE` | `6439474beed8e0271df9862eff97ffd70ec2464c` |

Upstream license, reproduced in full:

```
MIT License

Copyright (c) 2025 opencode

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

The port is by hand against the new contract, never spliced by script.
