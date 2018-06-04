package main

import (
  "fmt"
  "os"
  "github.com/InfinityTools/go-cmdargs"
)

func main() {
  // This will initialize an empty Parameter processor
  parameters := cmdargs.Create()

  // The following options should be recognized and evaluated by out example code:
  // Specifying double or single hyphen prefixes is optional.
  // A simple option with a single alias and no additional arguments. Parameter names and aliases are stored 
  // case-sensitive, but without hyphens internally. So "--A" and "-A" are treated as identical options, but "-A" 
  // and "-a" are not.
  parameters.AddParameter("help", []string{"h"}, 0)
  // An option without alias that requires one additional argument. Arguments for options are stored as Generic 
  // types internally. You can convert them to various basic datatypes by the available interface functions.
  // To convert a Generic to a string, use either String(), for additional type checking, or ToString().
  parameters.AddParameter("prefix", nil, 1)
  // An option with two aliases that requires one additional argument. To convert the Generic argument into a 
  // numeric datatype, use Int() to return a signed int or Uint() to return an unsigned int.
  // Both function return the converted value and a type check. Use ToInt() or ToUint() to skip the type check.
  // Conversion supports the following notations: decimal, hexadecimal (with "0x" prefix) and octal (with "0" prefix).
  parameters.AddParameter("num-threads", []string{"t", "T"}, 1)
  // An option that requires two additional arguments.
  parameters.AddParameter("position", nil, 2)

  // Now we parse our command line arguments
  if err := parameters.Evaluate(os.Args); err != nil {
    fmt.Printf("Evaluation error: %v\n", err)
    return
  }

  // Checking our options
  if parameters.GetArgExists("help") {
    // Print help and exit
    fmt.Printf("Usage: %s [--help] [--prefix path] [--num-threads n] [--position x y]\n", parameters.GetArgSelf())
    return
  }

  // Printing application name (if available)
  fmt.Printf("Application name: %s\n", parameters.GetArgSelf())

  // Getting additional parameter of --prefix. (Again, hyphen prefix is not strictly needed.)
  // We are force-converting the Generic into a string with ToString().
  // Note: For single argument options, the argument can also be specified as --option=argument.
  var prefix string
  if value, exists := parameters.GetArgParam("--prefix", 0); exists {
    prefix = value.ToString()
    fmt.Printf("Prefix is: %s\n", prefix)
  } else {
    prefix = "/my/default/prefix"
    fmt.Println("Prefix not specified. Using defaults.")
  }

  // Getting evaluated numeric argument via aliased option name.
  // Generic is converted via safe option Int().
  var numThreads int
  if value, exists := parameters.GetArgParam("t", 0); exists {
    if i, ok := value.Int(); ok && i > 0 {
      numThreads = int(i)
      fmt.Printf("Number of threads: %d\n", numThreads)
    } else {
      fmt.Printf("Error: Illegal number of threads: %v\n", value)
      return
    }
  } else {
    fmt.Println("Num. threads not specified. Using defaults.")
    numThreads = 1
  }

  // Getting multiple arguments of floating point type for option "position".
  var x, y float64 = 0, 0
  for idx := 0; idx < 2; idx++ {
    if value, exists := parameters.GetArgParam("position", idx); exists {
      if f, ok := value.Float(); ok {
        if idx == 0 {
          x = f
        } else {
          y = f
        }
      } else {
        fmt.Printf("Error: Invalid argument %d for option 'position': %v. Using default value 0.\n", idx, value)
      }
    }
  }
  fmt.Printf("Position: (%v, %v)\n", x, y)

  // Printing remaining extra arguments, such as filenames, etc.
  for idx, size := 0, parameters.GetArgExtraLength(); idx < size; idx++ {
    arg := parameters.GetArgExtra(idx)
    fmt.Printf("Extra arg #%d: %s\n", idx, arg.ToString())
    // Alternatively you could perform wildcard argument expansion:
    // list := parameters.GetExpandedArgExtra(idx)
    // fmt.Printf("Extra arg #%d (expanded): %v\n", idx, list)
  }
}
