package slicealiasdemo

// ───────────────────────────────────────────────────────────
//  1. Basic nil-slice aliasing (root = a, alias = b)
//     Expect a single warning on b’s append.
//
// ───────────────────────────────────────────────────────────
func basicNilAliasing() {
	var a []int // a is nil → unknown capacity
	b := a      // b aliases a

	a = append(a, 1) // root append: SAFE / ignore
	b = append(b, 2) // want "append to alias 'b' of unknown-capacity slice 'a'"
}

// ───────────────────────────────────────────────────────────
// 2. Empty literal aliasing ([]T{})
// ───────────────────────────────────────────────────────────
func literalAliasing() {
	a := []int{} // empty → unknown cap
	b := a

	b = append(b, 1) // want "append to alias 'b' of unknown-capacity slice 'a'"
	a = append(a, 2) // root append: SAFE / ignore
}

// ───────────────────────────────────────────────────────────
// 3. make(len=0) aliasing
// ───────────────────────────────────────────────────────────
func makeZeroLenAliasing() {
	a := make([]int, 0) // unknown cap
	b := a
	b = append(b, 1) // want "append to alias 'b' of unknown-capacity slice 'a'"
	a = append(a, 2) // ignore
}

// ───────────────────────────────────────────────────────────
// 4. make(len=0, cap>0)  → safe, no warnings
// ───────────────────────────────────────────────────────────
func makeSafeCap() {
	a := make([]int, 0, 100) // known cap
	b := a
	a = append(a, 1) // no warning
	b = append(b, 2) // no warning
}

// ───────────────────────────────────────────────────────────
//  5. Chain aliasing (c → b → a)
//     Only c’s append should warn; root append ignored
//
// ───────────────────────────────────────────────────────────
func chainAliasing() {
	var a []int
	b := a
	c := b

	a = append(a, 1) // root append: ignore
	c = append(c, 2) // want "append to alias 'c' of unknown-capacity slice 'a'"
}

// ───────────────────────────────────────────────────────────
// 6. Reassignment breaks alias; second section should be safe
// ───────────────────────────────────────────────────────────
func reassignmentBreak() {
	a := []int{}
	b := a
	b = append(b, 1) // want "append to alias 'b' of unknown-capacity slice 'a'"
	a = append(a, 2)

	// break alias chain
	a = []int{42}     // a now has its own backing array (len=1, cap=1)
	a = append(a, 99) // no warning
}

// ───────────────────────────────────────────────────────────
// 7. Non-alias append (single slice) – should never warn
// ───────────────────────────────────────────────────────────
func singleSlice() {
	a := make([]int, 0)
	a = append(a, 1) // no warning
	a = append(a, 2) // no warning
}
