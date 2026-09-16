# Attribution

`lasuite_ui` is an independent Flutter/Dart implementation of the design
system published by the **La Suite numérique UI Kit**:

- Source of truth: https://github.com/suitenumerique/ui-kit
- Public documentation: https://suitenumerique.github.io/ui-kit/
- Upstream license: MIT, Copyright (c) 2025 La Suite numérique

## What was translated

The design **tokens** (color palettes, semantic light/dark color mappings,
spacing scale, typography scale, border radii, motion durations/easing,
breakpoints) and the **component inventory, variants and states** (e.g.
Button's `primary/secondary/tertiary/bordered` variants and
`brand/neutral/info/success/warning/error` colors, Badge types, Alert types,
Field states, Modal sizes, etc.) were read from the upstream repository's
token engine (`@gouvfr-lasuite/ui-tokens`) and component source
(`@gouvfr-lasuite/ui-components`) to reproduce the same visual language and
interaction vocabulary.

No React/TypeScript/SCSS source code was copied. Every Dart widget in this
package is a fresh, idiomatic Flutter implementation built against Flutter's
own APIs (Theme/ThemeExtension, Material widgets, gestures, semantics).

## Differences from the upstream kit

See the "Differences from the original UI Kit" section of [README.md](README.md)
for font, icon and component-coverage differences.

## Upstream license

```
MIT License

Copyright (c) 2025 La Suite numérique

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
