import std/[os, strutils, sequtils]
import ../generators/default

# ==================== parseOne ====================
proc parseOne(line: string): (string, seq[string]) =
  # Skip lines that look like they're in documentation comments
  if line.strip().startsWith("*") or line.strip().startsWith("/*"):
    return ("", @[])
  # Find position of @
  let at = line.find('@')
  if at < 0:
    return ("", @[])

  # Check if there's // before @ on the same line
  let beforeAt = line[0 ..< at]
  if "//" in beforeAt:
    return ("", @[])

  # Note: We DON'T check for /* here because block comments
  # are handled in the main loop

  # Rest of parsing...
  let s = line.find('(', at)
  if s < 0:
    return ("", @[])

  let e = line.find(')', s)
  if e < 0:
    return ("", @[])

  let cmd = line[at + 1 ..< s].strip()
  let argsStr = line[s + 1 ..< e].strip()
  let args =
    if argsStr.len > 0:
      argsStr.split(',').mapIt(it.strip())
    else:
      @[]

  return (cmd, args)

# Helper to check if inside string/char literal
proc inStringLiteral(line: string, pos: int): bool =
  var inString = false
  var inChar = false
  var escaped = false

  for i in 0 ..< min(pos, line.len):
    if escaped:
      escaped = false
      continue

    case line[i]
    of '\\':
      escaped = true
    of '"':
      if not inChar:
        inString = not inString
    of '\'':
      if not inString:
        inChar = not inChar
    else:
      discard

  return inString or inChar

# ==================== main ====================
proc main() =
  const
    inputFile = "input/main.cccp"
    outputFile = "output/main.c"

  var
    content = readFile(inputFile)
    pos = 0
    inBlockComment = false

  while true:
    let start = content.find('@', pos)
    if start == -1:
      break

    # Check if this @ is inside a block comment
    if inBlockComment:
      var lineEnd = start
      while lineEnd < content.len and content[lineEnd] notin {'\n'}:
        inc lineEnd
      pos = lineEnd + 1
      continue

    # Find start of this line
    var lineStart = start
    while lineStart > 0 and content[lineStart - 1] notin {'\n'}:
      dec lineStart

    # Find end of this line  
    var lineEnd = start
    while lineEnd < content.len and content[lineEnd] notin {'\n'}:
      inc lineEnd

    # Get the full line
    let line = content[lineStart ..< lineEnd]

    # Update block comment state for this ENTIRE line
    var i = 0
    while i < line.len:
      # Skip if inside string/char literal
      if inStringLiteral(line, i):
        inc i
        continue

      if i + 1 < line.len:
        if line[i] == '/' and line[i + 1] == '*':
          inBlockComment = true
          i += 2
          continue
        elif line[i] == '*' and line[i + 1] == '/':
          inBlockComment = false
          i += 2
          continue

      inc i

    # Skip if we're now inside a block comment
    if inBlockComment:
      pos = lineEnd + 1
      continue

    # Parse the command
    let (cmdName, argList) = parseOne(line)

    # Check if we got a valid command
    if cmdName.len == 0:
      pos = lineEnd + 1
      continue

    # Generate replacement based on command
    var replacement = ""
    case cmdName
    of "getmem":
      replacement = malloc(argList)
    of "free":
      replacement = free(argList)
    of "stringcpy":
      replacement = stringcpy(argList)
    of "check":
      if argList.len == 0:
        pos = lineEnd + 1
        continue
      replacement = check(argList)
    of "moremem":
      if argList.len == 0:
        pos = lineEnd + 1
        continue
      replacement = moremem(argList)
    else:
      pos = lineEnd + 1
      continue

    # Replace from line start to line end
    content = content[0 ..< lineStart] & replacement & content[lineEnd ..^ 1]
    pos = lineStart + replacement.len

  writeFile(outputFile, content)
  discard execShellCmd("clang-format -i " & outputFile & " 2>/dev/null")

when isMainModule:
  main()
