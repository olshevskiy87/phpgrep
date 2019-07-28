package phpgrep

import (
	"testing"
)

type matcherTest struct {
	pattern string
	input   string
}

func mustCompile(t *testing.T, c *Compiler, code string) *Matcher {
	matcher, err := c.Compile([]byte(code))
	if err != nil {
		t.Fatalf("pattern compilation error:\ntext: %q\nerr: %v", code, err)
	}
	return matcher
}

func TestFind(t *testing.T) {
	runFindTest := func(t *testing.T, pattern, code string, wantMatches []string) {
		var c Compiler
		matcher := mustCompile(t, &c, pattern)
		var haveMatches []string
		matcher.Find([]byte(code), func(m *MatchData) bool {
			haveMatches = append(haveMatches, string(code[m.PosFrom:m.PosTo]))
			return true
		})
		if len(haveMatches) != len(wantMatches) {
			t.Errorf("matches count mismatch:\nhave: %d\nwant: %d",
				len(haveMatches), len(wantMatches))
			return
		}
		for i, have := range haveMatches {
			want := wantMatches[i]
			if have != want {
				t.Errorf("match mismatch:\nhave: %q\nwant: %q", have, want)
			}
		}
	}

	runFindTest(t, `$x+1`, `<?php $x+1;`, []string{`$x+1`})

	runFindTest(t, `$x = $x`, `<?php
            $x = $x; $z1 = 10; $y = $y; $z2 = 20;
        `, []string{
		`$x = $x`,
		`$y = $y`,
	})
}

func runMatchTest(t *testing.T, want bool, tests []*matcherTest) {
	var c Compiler
	for _, test := range tests {
		matcher := mustCompile(t, &c, test.pattern)

		have := matcher.Match([]byte(test.input))
		if have != want {
			t.Errorf("match results mismatch:\npattern: %q\ninput: %q\nhave: %v\nwant: %v",
				test.pattern, test.input, have, want)
		}
	}
}

func TestMatchDebug(t *testing.T) {
	runMatchTest(t, true, []*matcherTest{
		{`if ($c) $_; else if ($c) {1;};`, `if ($c1) {1; 2;} else if ($c1) {1;}`},
	})
}

func TestMatch(t *testing.T) {
	runMatchTest(t, true, []*matcherTest{
		{`$x=$x`, `$x=$x`},

		{`1`, `1`},
		{`"1"`, `"1"`},
		{`'1'`, `'1'`},
		{`1.4`, `1.4`},

		{`$x & mask != 0`, `$v & (mask != 0)`},
		{`($x & mask) != 0`, `($v & mask) != 0`},

		{`$x`, `10`},
		{`$x`, `"abc"`},
		{`false`, `false`},
		{`NULL`, `NULL`},

		{`$x++`, `$y++`},
		{`$x--`, `$y--`},
		{`++$x`, `++$y`},
		{`--$x`, `--$y`},

		{`$x+1`, `10+1`},
		{`$x+1`, `$x+1`},
		{`$x-1`, `10-1`},
		{`$x-1`, `$x-1`},

		{`+$x`, `+1`},
		{`-$x`, `-2`},
		{`~$x`, `~$v`},
		{`!$x`, `!$v`},

		{`$f()`, `f()`},
		{`$f()`, `g()`},
		{`$f($a1, $a2)`, `f(1, 2)`},
		{`$f($a1, $a2)`, `f("sa", $t)`},

		{`$x + $x`, `1 + 1`},
		{`$x + $y`, `1 + 1`},
		{`$x | $y`, `$v1 | $v2`},
		{`$x >> $y`, `$v1 >> $v2`},
		{`$x << $y`, `$v1 << $v2`},
		{`$x and $y`, `$v1 and $v2`},
		{`$x or $y`, `$v1 or $v2`},
		{`$x xor $y`, `$v1 xor $v2`},
		{`$x != $y`, `$v1 != $v2`},
		{`$x == $y`, `$v1 == $v2`},
		{`$x === $y`, `$v1 === $v2`},
		{`$x !== $y`, `$v1 !== $v2`},
		{`$x > $y`, `$v1 > $v2`},
		{`$x >= $y`, `$v1 >= $v2`},
		{`$x < $y`, `$v1 < $v2`},
		{`$x <= $y`, `$v1 <= $v2`},
		{`$x <=> $y`, `$v1 <=> $v2`},
		{`$x && $y`, `$v1 && $v2`},
		{`$x || $y`, `$v1 || $v2`},
		{`$x ?? $y`, `$v1 ?? $v2`},
		{`$x . $y`, `$v1 . $v2`},
		{`$x / $y`, `$v1 / $v2`},
		{`$x % $y`, `$v1 % $v2`},
		{`$x * $y`, `$v1 * $v2`},
		{`$x ** $y`, `$v1 ** $v2`},

		{`int($x)`, `int($v)`},
		{`array($x)`, `array($v)`},
		{`string($x)`, `string($v)`},
		{`bool($x)`, `bool($v)`},
		{`double($x)`, `double($v)`},
		{`object($x)`, `object($v)`},

		{`$$$x`, `$$x`},
		{`$$$x`, `$$y`},

		{`$x = 0`, `$v = 0`},
		{`$x += 1`, `$v += 1`},
		{`$x -= 1`, `$v -= 1`},
		{`$x =& $y`, `$x =& $y`},
		{`$x &= $y`, `$x &= $y`},
		{`$x |= $y`, `$x |= $y`},
		{`$x ^= $y`, `$x ^= $y`},
		{`$x /= $y`, `$x /= $y`},
		{`$x %= $y`, `$x %= $y`},
		{`$x *= $y`, `$x *= $y`},
		{`$x **= $y`, `$x **= $y`},
		{`$x <<= $y`, `$x <<= $y`},
		{`$x >>= $y`, `$x >>= $y`},

		{`\A\B`, `\A\B`},

		{`[]`, `[]`},
		{`array()`, `array()`},
		{`[$x, $x]`, `[1, 1]`},
		{`array($x, $x)`, `array(1, 1)`},
		{`[$k1 => 2, $k2 => 4]`, `[1 => 2, 3 => 4]`},
		{`[$k1 => 2, $k1 => 4]`, `[1 => 2, 1 => 4]`},

		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[1 => $x, 1 => $y]`},
		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[$v, 1 => $x, $v, 1 => $x, $v]`},
		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[1 => $x, 1 => $x, $v]`},
		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[$v, 1 => $x, 1 => $x]`},

		{`{1; 2;}`, `{1; 2;}`},
		{`{$x;}`, `{1;}`},
		{`{$x;}`, `{2;}`},

		{`{${'*'};}`, `{}`},
		{`{${'*'};}`, `{1;}`},
		{`{${'*'};}`, `{1; 2;}`},
		{`{${'*'};}`, `{1; 2; 3;}`},

		{`{${'*'}; 3;}`, `{1; 2; 3;}`},
		{`{1; ${'*'};}`, `{1; 2; 3;}`},
		{`{1; ${'*'}; 3;}`, `{1; 2; 3;}`},
		{`{1; 2; ${'*'}; 3;}`, `{1; 2; 3;}`},
		{`{${'*'}; 2; ${'*'};}`, `{1; 2; 3;}`},
		{`{1; 2; 3; ${'*'};}`, `{1; 2; 3;}`},

		{`f(${'*'})`, `f()`},
		{`f(${'*'})`, `f(1)`},
		{`f(${'*'})`, `f(1, 2)`},
		{`f(${'*'})`, `f(1, 2, 3)`},
		{`f(${'*'}, 3)`, `f(1, 2, 3)`},
		{`f(${'*'}, $x, $y, $z)`, `f(1, 2, 3)`},
		{`f($x, $y, $z, ${'*'})`, `f(1, 2, 3)`},
		{`f(${'*'}, $x, ${'*'}, $y, ${'*'}, $z, ${'*'})`, `f(1, 2, 3)`},

		{`if ($cond) $_;`, `if (1 == 1) return 1;`},
		{`if ($cond) $_;`, `if (1 == 1) f();`},
		{`if ($cond) return 1;`, `if (1 == 1) return 1;`},
		{`if ($cond) { return 1; }`, `if (1 == 1) { return 1; }`},
		{`if ($_ = $_) $_`, `if ($x = f()) {}`},
		{`if ($_ = $_) $_`, `if ($x = f()) g();`},
		{`if ($cond1) $_; else if ($cond2) $_;`, `if ($c1) {} else if ($c2) {}`},
		{`if ($cond1) $_; elseif ($cond2) $_;`, `if ($c1) {} elseif ($c2) {}`},

		{`switch ($e) {}`, `switch ($x) {}`},
		{`switch ($_) {case 1: f();}`, `switch ($x) {case 1: f();}`},
		{`switch ($_) {case $_: ${'*'};}`, `switch ($x) {case 1: f1(); f2();}`},
		{`switch ($e) {default: $_;}`, `switch ($x) {default: 1;}`},

		{`strcmp($s1, $s2) > 0`, `strcmp($s1, "foo") > 0`},

		{`new $t`, `new T`},
		{`new $t()`, `new T()`},
		{`new $t($x)`, `new T(1)`},
		{`new $t($x, $y)`, `new T(1, 2)`},
		{`new $t(${'*'})`, `new T(1, 2)`},

		{`list($x, $_, $x) = f()`, `list($v, $_, $v) = f()`},
		{`list($x, $_, $x) = f()`, `list($v, , $v) = f()`},
		{`list($x) = $a`, `list($v) = [1]`},

		{`${'var'}`, `$x`},
		{`${'var'}`, `$$x`},
		{`${'x:var'} + $x`, `$x + $x`},
		{`$x + ${'x:var'}`, `$x + $x`},
		{`${'_:var'} + $_`, `$x + 1`},
		{`${'var'} + $_`, `$x + 1`},

		{`${"int"}`, `13`},
		{`${"float"}`, `3.4`},
		{`${"str"}`, `"123"`},
		{`${"num"}`, `13`},
		{`${"num"}`, `3.4`},

		{`${"expr"}`, `1`},
		{`${"expr"}`, `"124d"`},
		{`${"expr"}`, `$x`},
		{`${"expr"}`, `f(1, 5)`},
		{`${"expr"}`, `$x = [1]`},

		{`$cond ? $true : $false`, `1 ? 2 : 3`},
		{`$cond ? a : b`, `1 ? a : b`},
		{`$c1 ? $_ : $_ ? $_ : $_`, `true ? 1 : false ? 2 : 3`},
		{`($c1 ? $_ : $_) ? $_ : $_`, `true ? 1 : false ? 2 : 3`},
		{`$c1 ? $_ : ($_ ? $_ : $_)`, `true ? 1 : (false ? 2 : 3)`},
		{`$x ? $x : $y`, `$v ? $v : $other`},

		{`$_ ?: $_`, `1 ?: 2`},

		{`isset($x)`, `isset($v)`},
		{`isset($x, $y)`, `isset($k, $v[$k])`},
		{`empty($x)`, `empty($v)`},

		{`$x->$_ = $x`, `$this->self = $this`},
		{`$x->$_ = $x`, `$this->$indirect = $this`},
		{`$x->$m()`, `$this->m()`},
		{`$x->$m(1, 2)`, `$this->m(1, 2)`},
		{`$x->ff(1, 2)`, `$this->ff(1, 2)`},

		{`$_[0]`, `$v[0]`},

		{`$c::$prop`, `C::$foo`},
		{`$c::$prop`, `C::constant`},
		{`$c::$f()`, `C::foo()`},
		{`$c::$f()`, `C::$foo()`},
		{`C::f()`, `C::f()`},
		{`C::constant`, `C::constant`},

		{`clone $v`, `clone new T()`},

		{`@$_`, `@f()`},
		{`@$_`, `@$o->method(1, 2)`},

		{`eval($_)`, `eval('1')`},

		{`exit(0)`, `exit(0)`},
		{`die(0)`, `die(0)`},

		{`include $_`, `include "foo.php"`},
		{`include_once $_`, `include_once "foo.php"`},
		{`require $_`, `require "foo.php"`},
		{`require_once $_`, `require_once "foo.php"`},

		{`__FILE__`, `__FILE__`},
		{`[$x, $x]`, `[__FILE__, __FILE__]`},

		{`"$x$y"`, `"$x$y"`},
		{`"$x 1" . $x`, `"$x 1" . "2"`},
		{`"${x}"`, `"${x}"`},
	})
}

func TestMatchNegative(t *testing.T) {
	runMatchTest(t, false, []*matcherTest{
		{`1`, `2`},
		{`"1"`, `"2ed"`},
		{`'1'`, `'x'`},
		{`1.4`, `1.6`},
		{`false`, `true`},

		{`$x+1`, `10+2`},
		{`$x+1`, `$x+$x`},

		{`+$x`, `-1`},
		{`-$x`, `+2`},

		{`$f()`, `f(1)`},
		{`$f()`, `g(2)`},
		{`$f($a1, $a2)`, `f()`},
		{`$f($a1, $a2)`, `f()`},

		{`$x+$x`, `1+2`},
		{`$x+$x`, `2+1`},
		{`$x+$x`, `""+1`},
		{`$x+$x`, `1+""`},

		{`$$$x`, `$x`},
		{`$$$x`, `10`},

		{`[$x, $x]`, `[1, 2]`},
		{`array($x, $x)`, `array(1, 2)`},

		{`{}`, `{1;}`},
		{`{1;}`, `{}`},
		{`{1;}`, `{1; 2;}`},
		{`{1; 2;}`, `{1; 2; 3;}`},
		{`{1; 2; 3;}`, `{1; 2;}`},

		{`f(${'*'}, 4)`, `f(1, 2, 3)`},

		{`new $t`, `new T()`},
		{`new $t()`, `new T`},

		{`while ($_); {${'*'};}`, `while ($cond) {$blah;}`},

		{`if ($c) $_; else if ($c) $_;`, `if ($c1) {} else if ($c2) {}`},
		{`if ($c) $_; elseif ($c) $_;`, `if ($c1) {} elseif ($c2) {}`},

		{`list($x, $_, $x) = f()`, `list(,1,2) = f()`},
		{`list($x, $_, $x) = f()`, `list(2,1,) = f()`},

		{`${'x:var'}`, `1`},
		{`${'var'}`, `[10]`},
		{`${'var'}`, `THE_CONST`},
		{`${'x:var'} + $x`, `$x + 1`},
		{`$x + ${'x:var'}`, `1 + $x`},

		{`${"int"}`, `13.5`},
		{`${"float"}`, `3`},
		{`${"str"}`, `5`},
		{`${"num"}`, `$x`},
		{`${"num"}`, `"1"`},

		{`${"expr"}`, `{}`},
		{`${"expr"}`, `{{}}`},

		{`$c1 ? $_ : $_ ? $_ : $_`, `true ? 1 : (false ? 2 : 3)`},

		{`$x ? $x : $y`, `1 ?: 2`},

		{`$x->$_ = $x`, `$this->self = $y`},

		{`$_[0]`, `$v[1]`},

		{`@$_`, `f()`},

		{`die(0)`, `exit(0)`},
		{`exit(0)`, `die(0)`},

		{`$x->$m()`, `$this->m(1)`},
		{`$x->$m(1, 2)`, `$this->m(2, 1)`},
		{`$x->ff(1, 2)`, `$this->f2(1, 2)`},

		{`C::f()`, `C2::f()`},
		{`C::f()`, `C::f2()`},
		{`C::constant`, `C::constant2`},
		{`C::constant`, `C::$prop`},

		{`__FILE__`, `__DIR__`},
		{`[$x, $x]`, `[__FILE__, __DIR__]`},
		{`[$x, $x]`, `[__DIR__, __FILE__]`},

		{`"$x$x"`, `"11"`},
		{`"${x}"`, `"$x"`},
		{`"$x$x"`, `'$x$x'`},

		{`int($x)`, `$v`},
		{`array($x)`, `$v`},
		{`string($x)`, `$v`},
		{`bool($x)`, `$v`},
		{`double($x)`, `$v`},
		{`object($x)`, `$v`},

		{`\A\B`, `\A\A`},
		{`\A\B`, `\B\B`},
	})
}
