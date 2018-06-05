/*
Package cmdargs implements a command line arguments parser.
*/
package cmdargs

import (
  "fmt"
  "errors"
  "path/filepath"
  "strings"
)

// Definition for a single parameter
type paramType struct {
  name      string      // Long name of the parameter (without prefix)
  numArgs   int         // Expected number of arguments.
}

// Storage for a single argument
type optionType struct {
  value     []Generic
  position  int
}

// Maps parameter/argument entries
type paramList []*paramType
type paramMap map[string]*paramType
type optionMap map[string]*optionType

type Parameter struct {
  definitions paramList   // sequence of parameter definitions
  aliases     paramMap    // map for parameter/alias names to parameter definitions

  options     optionMap   // map for parameter names to evaluated command line options
  extra       []Generic   // Remaining list of unparsed command line arguments (e.g. file names, etc.)
  self        string      // Contains the application name (args[0]), unless it is identified as an option.
}


// Create creates an empty Parameter structure
func Create() *Parameter {
  p := Parameter { definitions: make(paramList, 0),
                   aliases: make(paramMap),
                   options: make(optionMap),
                   extra: make([]Generic, 0),
                   self: "" }
  return &p
}

// AddParameter adds or updates the parameter definition specified by "name".
//
// name is the primary name of the parameter, e.g. "--help". Prefix "-" or "--" may be omitted.
// aliases is a sequence of alternate names for the parameter. Alias prefixes may also be omitted.
// Specify nil or an empty array to skip.
// numArgs specifies the number of additional arguments belonging to the parameter.
//
// Note:
// By convention, long option names start with two hyphens and single character options start with a single hyphen.
// Parameter names and aliases are stored internally case-sensitive and with their prefix stripped.
// For example, "--A" and "-A" will override each other, but "-A" and "-a" will not.
func (param *Parameter) AddParameter(name string, aliases []string, numArgs int) {
  name = getOptionName(name)
  if len(name) == 0 { return }
  if aliases != nil {
    for i, alias := range aliases {
      aliases[i] = getOptionName(alias)
    }
  }
  if numArgs < 0 { numArgs = 0 }

  p, ok := param.aliases[name]
  if !ok {
    p = &paramType{name: name, numArgs: 0}
    param.definitions = append(param.definitions, p)
  }
  p.numArgs = numArgs

  param.aliases[name] = p
  if aliases != nil {
    for _, a := range aliases {
      if len(a) > 0 {
        param.aliases[a] = p
      }
    }
  }
}

// RemoveParameter removes the parameter of given name. Returns whether there was a parameter definition that could
// be removed.
func (param *Parameter) RemoveParameter(name string) bool {
  name = getOptionName(name)
  p, ok := param.aliases[name]
  if ok {
    // removing definition object from array
    for idx := 0; idx < len(param.definitions); idx++ {
      if param.definitions[idx] == p {
        if idx+1 < len(param.definitions) {
          copy(param.definitions[idx:], param.definitions[idx+1:])
        }
        param.definitions = param.definitions[:len(param.definitions)-1]
        break
      }
    }
    // removing definition references from alias map
    for alias, def := range param.aliases {
      if def == p {
        delete(param.aliases, alias)
      }
    }
  }
  return ok
}

// Evaluate parses and evaluates the arguments in the given string array, so that they can be directly accessed by
// the respective argument functions.
//
// Parameter evaluation stops at the first occurence of a non-parameter string.
// Remaining entries will be stored as an unparsed list of extra arguments. First entry will be stored as application
// name, unless it is identified as an option.
// Multiple instances of the same parameter (specified either by name or alias) will override each other. Only the last
// instance will be preserved.
//
// Returns an error if a parameter is found that doesn't match any entries added by AddParameter.
func (param *Parameter) Evaluate(args []string) error {
  var err error = nil
  if args == nil || len(args) == 0 { return err }

  param.reset()
  argIdx := 0

  // initializing "self"
  if !isOption(args[argIdx]) {
    param.self = args[argIdx]
    argIdx++
  }

  // parsing options
  position := 0
  for argIdx < len(args) {
    var name string
    var arg *optionType
    oldIdx := argIdx
    name, arg, argIdx, err = param.evalArg(args, argIdx)
    if err != nil { return err }
    if name == "" { break }     // remaining entries are not options
    if argIdx == oldIdx { return errors.New("Fatal: Deadlock while evaluating parameters") }  // should never happen!
    name = getOptionName(name)  // normalizing option name
    arg.position = position
    if oldArg, ok := param.options[name]; ok {
      oldArg.value = arg.value
    } else {
      param.options[name] = arg
    }
    position++
  }

  // initializing extra arguments
  for idx := argIdx; idx < len(args); idx++ {
    param.extra = append(param.extra, Generic(String(args[idx])))
  }

  return err
}


// GetArgSelf returns the first argument of the argument list, unless it was identified as a regular option.
// It is usually the application name.
func (param *Parameter) GetArgSelf() string {
  return param.self
}


// GetArgExtraLength returns the number of available extra arguments that were not evaluated as regular options.
func (param *Parameter) GetArgExtraLength() int {
  return len(param.extra)
}

// GetArgExtra returns the extra argument at the specified index.
func (param *Parameter) GetArgExtra(index int) Generic {
  if index < 0 || index > param.GetArgExtraLength() { return String("") }
  return param.extra[index]
}

// GetExpandedArgExtra treats the given argument as a wildcard and expands it relative to current directory.
//
// Returns a list with all matching path strings (which may be empty if no match is found), or an empty list on error.
func (param *Parameter) GetExpandedArgExtra(index int) []string {
  retVal := make([]string, 0)
  if index < 0 || index > param.GetArgExtraLength() { return retVal }

  expanded, err := filepath.Glob(param.extra[index].ToString())
  if err == nil {
    retVal = append(retVal, expanded...)
  }
  return retVal
}


// GetArgLength returns the number of evaluated options.
func (param *Parameter) GetArgLength() int {
  return len(param.options)
}

// GetArgExists returns whether the argument of given name has been evaluated by a previous
// call to Evaluate. It considers option names and aliases.
func (param *Parameter) GetArgExists(name string) bool {
  exists := false
  name = param.getLongOptionName(name)
  if len(name) > 0 {
    _, exists = param.options[name]
  }
  return exists
}

// GetArgPosition returns the position of the specified option in the command line options list.
func (param *Parameter) GetArgPosition(name string) (pos int, exists bool) {
  name = param.getLongOptionName(name)
  if len(name) > 0 {
    arg, exists := param.options[name]
    if exists {
      pos = arg.position
    }
  }
  return
}

// GetArgNameByPosition returns the name of the option at the specified position.
func (param *Parameter) GetArgNameByPosition(pos int) (name string, exists bool) {
  if pos < 0 || pos >= len(param.options) { return }
  for key, arg := range param.options {
    if arg.position == pos {
      name = key
      exists = true
      break
    }
  }
  return
}

// GetArgParamLength returns the number of additional arguments available for the specified option.
//
// Options with a variable number of arguments will return the actually parsed number of arguments.
// Function returns -1 for non-existing options.
func (param *Parameter) GetArgParamLength(name string) int {
  name = param.getLongOptionName(name)
  if len(name) == 0 { return -1 }
  arg, exists := param.options[name]
  if !exists { return -1 }
  return len(arg.value)
}

// GetArgParam returns the argument of the given name at specified index.
//
// value contains the returned argument value. exists returns whether the value (either option or argument) exists.
func (param *Parameter) GetArgParam(name string, index int) (value Generic, exists bool) {
  name = param.getLongOptionName(name)
  if len(name) == 0 { return }

  var arg *optionType
  arg, exists = param.options[name]
  if !exists { return }
  if index < 0 || index >= len(arg.value) { exists = false; return }
  value = arg.value[index]
  return
}


// Used internally. Resets all Parameter fields that are related to argument evaluation to initial state.
func (param *Parameter) reset() {
  if len(param.options) != 0 {
    param.options = make(optionMap)
  }
  if len(param.extra) != 0 {
    param.extra = make([]Generic, 0)
  }
  param.self = ""
}

// Used internally. Attempts to parse the next available command line argument.
func (param *Parameter) evalArg(args []string, index int) (name string, arg *optionType, newIdx int, err error) {
  arg = &optionType{value: make([]Generic, 0), position: -1} // position is set by the calling function
  newIdx = index
  if newIdx < 0 || newIdx >= len(args) { return }

  // remaining arguments are treated as non-options
  if !isOption(args[newIdx]) { return }

  // parsing new option
  args0 := strings.Split(args[newIdx], "=")
  if len(args0) == 0 { return }
  name = getOptionName(args0[0])
  newIdx++

  def, ok := param.aliases[name]
  if !ok { err = fmt.Errorf("Unrecognized option: \"--%s\" or \"-%s\"", name, name); return }

  numArgs := def.numArgs

  // option may contain extra argument, separated by equal sign
  if numArgs > 0 && len(args0) > 1 {
    s := args0[1]
    for i := 2; i < len(args0); i++ {
      // subsequent equal signs are treated as part of the extra argument
      s += "="
      s += args[i]
    }
    arg.value = append(arg.value, trimArg(s))
    numArgs--
  }

  numRemaining := len(args) - newIdx
  if numRemaining < numArgs {
    err = fmt.Errorf("Too few option arguments: available=%d, need=%d", numRemaining, numArgs);
    return
  }

  // parsing remaining option arguments
  for ; numArgs > 0; numArgs, newIdx = numArgs-1, newIdx+1 {
    a := args[newIdx]
    arg.value = append(arg.value, trimArg(a))
  }

  name = param.getLongOptionName(name)  // always returns long option name

  return
}

// Used internally. Removes spaces and double-quotes from arguments if needed.
func trimArg(arg string) Generic {
  // strip double-quotes from arguments, but leave single quotes unchanged
  arg = strings.TrimSpace(arg)
  if len(arg) > 0 {
    if arg[0] == '"' && arg[len(arg) - 1] == '"' {
      arg = arg[1:]
      if len(arg) > 0 && arg[len(arg) - 1] == '"' {
        arg = arg[:len(arg) - 1]
      }
    }
  }
  return String(arg)
}

// Used internally. Returns the long name of the parameter referenced by the given alias.
func (param *Parameter) getLongOptionName(alias string) string {
  retVal := ""
  alias = getOptionName(alias)
  if len(alias) > 0 {
    if p, ok := param.aliases[alias]; ok {
      retVal = p.name
    }
  }
  return retVal
}

// Used internally. Strips prefix from option name.
func getOptionName(name string) string {
  if len(name) >= 2 && name[:2] == "--" {
    return name[2:]
  } else if len(name) >= 1 && name[:1] == "-" {
    return name[1:]
  } else {
    return name
  }
}

// Used internally. Returns whether the argument qualifies as an option name.
func isOption(name string) bool {
  return (len(name) >= 2 && name[:2] == "--") || (len(name) >= 1 && name[:1] == "-")
}
