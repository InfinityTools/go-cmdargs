package cmdargs

import (
  "strconv"
)

// Generic provides a set of methods that can be used to convert the value into specific types.
type Generic interface {
  String() (string, bool)
  ToString() string
  Bool() (bool, bool)
  ToBool() bool
  Int() (int64, bool)
  ToInt() int64
  Uint() (uint64, bool)
  ToUint() uint64
  Float() (float64, bool)
  ToFloat() float64
}

// Underlying type that implements the Generic interface.
type String string


// String simply returns the unaltered string value of the String datatype.
func (t String) String() (ret string, ok bool) {
  ret = string(t)
  ok = true
  return
}

// ToString behaves just like String, but omits the second return value. Returns the zero value of type string in
// case of an error.
func (t String) ToString() string {
  ret, _ := t.String()
  return ret
}

// Bool returns true for strings "t", "T", "TRUE", "true", "True" and any non-zero numeric values.
// It returns false for "f", "F", "FALSE", "false", "False" and numeric zero.
// ok indicates whether the conversion was successful.
func (t String) Bool() (ret bool, ok bool) {
  b, err := strconv.ParseBool(string(t))
  if err == nil { ret = b; ok = true; return }

  i, err := strconv.ParseInt(string(t), 0, 0)
  if err == nil { ret = (i != 0); ok = true; return }

  f, err := strconv.ParseFloat(string(t), 64)
  if err == nil { ret = (f != 0.0); ok = true; return }

  return
}

// ToBool behaves just like Bool, but omits the second return value. Returns the zero value of type bool in case of
// an error.
func (t String) ToBool() bool {
  ret, _ := t.Bool()
  return ret
}

// Int attempts to interpret the string as a numeric value. It takes prefixes into account to determine
// the right numeric base. Boolean strings will be converted to 0 for "false" and 1 for "true".
func (t String) Int() (ret int64, ok bool) {
  i, err := strconv.ParseInt(string(t), 0, 64)
  if err == nil { ret = i; ok = true; return }

  f, err := strconv.ParseFloat(string(t), 64)
  if err == nil { ret = int64(f); ok = true; return }

  b, err := strconv.ParseBool(string(t))
  if err == nil { if b { ret = 1 }; ok = true; return }

  return
}

// ToInt behaves just like Int, but omits the second return value. Returns the zero value of type int64 in case of
// an error.
func (t String) ToInt() int64 {
  ret, _ := t.Int()
  return ret
}

// Uint attempts to interpret the string as an unsigned numeric value. It takes prefixes into account to determine
// the right numeric base. Boolean strings will be converted to 0 for "false" and 1 for "true".
func (t String) Uint() (ret uint64, ok bool) {
  u, err := strconv.ParseUint(string(t), 0, 64)
  if err == nil { ret = u; ok = true; return }

  f, err := strconv.ParseFloat(string(t), 64)
  if err == nil && f >= 0.0 { ret = uint64(f); ok = true; return }

  b, err := strconv.ParseBool(string(t))
  if err == nil { if b { ret = 1 }; ok = true; return }

  return
}

// ToUint behaves just like Uint, but omits the second return value. Returns the zero value of type uint64 in case of
// an error.
func (t String) ToUint() uint64 {
  ret, _ := t.Uint()
  return ret
}

// Float attempts to interpret the string as a floating point value.
// Boolean strings will be converted to 0 for "false" and 1 for "true".
func (t String) Float() (ret float64, ok bool) {
  f, err := strconv.ParseFloat(string(t), 64)
  if err == nil { ret = f; ok = true; return }

  i, err := strconv.ParseInt(string(t), 0, 64)
  if err == nil { ret = float64(i); ok = true; return }

  b, err := strconv.ParseBool(string(t))
  if err == nil { if b { ret = 1.0 }; ok = true; return }

  return
}

// ToFloat behaves just like Float, but omits the second return value. Returns the zero value of type float64 in case
// of an error.
func (t String) ToFloat() float64 {
  ret, _ := t.Float()
  return ret
}
