# Review notes

- Keep out-of-range clearing tolerant in buffer fast paths. `SetCell` ignores
  missing rows and cells; a row-slice optimization must not turn a clearing
  operation that previously succeeded into a slice-bounds panic.
