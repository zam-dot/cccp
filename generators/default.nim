import std/[strutils, strformat]

# ==================== malloc ====================
# Usage:
# @getmem(int, ptr, size);
proc malloc*(args: seq[string]): string =
  let argList =
    fmt"""
{args[0]} *{args[1]} = malloc({args[2]} * sizeof({args[0]}));
   if ({args[1]} == NULL) {{
        perror("malloc failed");
        exit(EXIT_FAILURE);
    }}"""
  return argList

# ==================== free ====================
# Usage:
# @free(ptr);
proc free*(args: seq[string]): string =
  let argList =
    fmt"""
free({args[0]});
{args[0]} = NULL;"""
  return argList

# ================== stringcpy ===================
# Usage:
# @stringcpy(dest, src);
proc stringcpy*(args: seq[string]): string =
  var actualSource = args[1]
  if actualSource.startsWith('"') and not actualSource.endsWith('"'):
    actualSource = args[1 ..^ 1].join(",")

  fmt"""{{
size_t src_len = strlen({actualSource});
size_t dest_size = sizeof({args[0]});
size_t copy_len = (src_len < dest_size - 1) ? src_len : dest_size - 1;
memcpy({args[0]}, {actualSource}, copy_len);
{args[0]}[copy_len] = '\0';
}}"""

# ==================== check ====================
# Usage:
# @check(ptr, "malloc failed");
proc check*(args: seq[string]): string =
  let argList =
    fmt"""
if ({args[0]} == NULL) {{
    perror({args[1]});
    exit(EXIT_FAILURE);
}}"""
  return argList

# ==================== moremem ====================
# Usage:
# @moremem(int, ptr, new_size);
proc moremem*(args: seq[string]): string =
  # @moremem(type, ptr, new_size)
  fmt"""
{{
    {args[0]} *_tmp_ptr = realloc({args[1]}, {args[2]} * sizeof({args[0]}));
    if (_tmp_ptr == NULL) {{
        perror("realloc failed");
        free({args[1]});
        exit(EXIT_FAILURE);
    }}
    {args[1]} = _tmp_ptr;
}}"""
