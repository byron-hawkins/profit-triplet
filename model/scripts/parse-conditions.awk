#!/bin/awk

BEGIN{ a = 0 }

/@-/ { a = 1; i = 0 }

/@\^/ {
  gsub(/^[^@]*@/, "@", $0)
  gsub(/:$/, "", $0)
  print ""
  print $0
  next
}

/\*\// { a = 0 }

{
  if (a) {
    if (i == 0)
      print ""
    else
      gsub(/\(\?\)/, "(" i ")", $0)

    gsub(/\t/, "", $0)
    gsub(/\/\*/, "", $0)
    gsub(/^ */, "", $0)
    gsub(/^\* /, "", $0)
    print $0

    i = i + 1
  }
}
