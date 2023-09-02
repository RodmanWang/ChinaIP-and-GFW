#!/bin/bash

# Function to calculate MASK
MASK() {
  x=$1
  if [ $x -ne 0 ]; then
    echo $((~((1<<(32-x))-1)))
  else
    echo 0
  fi
}

current=0
root=""

# Struct for Trie
struct_Trie() {
  flag=0
  child=()
}

# Function to merge Trie nodes
merge() {
  p=$1
  if [ "$flag" -eq 1 ]; then
    return 1
  fi

  if [ -z "${p.child[0]}" ] || [ -z "${p.child[1]}" ]; then
    return 0
  fi

  if [ "$(merge ${p.child[0]})" -eq 1 ] && [ "$(merge ${p.child[1]})" -eq 1 ]; then
    p.flag=1
    return 1
  fi

  return 0
}

# Function to print merged subnets
print() {
  p=$1
  depth=$2

  if [ "$p.flag" -eq 1 ]; then
    ip=$((current & $(MASK $depth)))
    echo "$((ip>>24&0xff)).$((ip>>16&0xff)).$((ip>>8&0xff)).$((ip&0xff))/$depth"
    return
  fi

  if [ -n "${p.child[0]}" ]; then
    current=$((current & ~(1<<(31-depth))))
    print "${p.child[0]}" $((depth+1))
  fi

  if [ -n "${p.child[1]}" ]; then
    current=$((current | 1<<(31-depth)))
    print "${p.child[1]}" $((depth+1))
  fi
}

# Main function
main() {
  while read -r ip; do
    IFS='/' read -ra parts <<< "$ip"
    IFS='.' read -ra ip_parts <<< "${parts[0]}"
    ip1=${ip_parts[0]}
    ip2=${ip_parts[1]}
    ip3=${ip_parts[2]}
    ip4=${ip_parts[3]}
    prefix_len=${parts[1]}

    ip=$((($ip1<<24) | ($ip2<<16) | ($ip3<<8) | $ip4))
    mask=$(MASK $prefix_len)

    p=root
    while [ "$mask" -ne 0 ]; do
      if [ -z "${p.child[ip>>31]}" ]; then
        p.child[ip>>31]=$(struct_Trie)
      fi
      p=${p.child[ip>>31]}
      ip=$((ip<<1))
      mask=$((mask<<1))
    done

    if [ -z "${p.child[0]}" ]; then
      p.child[0]=$(struct_Trie)
    fi
    p=${p.child[0]}
    p.flag=1
  done < dist/ChinaIPv4v6_tamp.txt > dist/ChinaIPv4v6.txt

  merge "$root"
  print "$root" 0
}

main
