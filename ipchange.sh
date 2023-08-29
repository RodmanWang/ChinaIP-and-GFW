#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <input_file>"
    exit 1
fi

input_file="$1"
output_file="dist1/ChinaIPv4v6.txt"

awk -F "/" '{
    ip = $1;
    mask = $2;
    type = "IPv6";

    if (match(ip, ":")) {
        type = "IPv6";
    } else {
        type = "IPv4";
    }

    if (type == "IPv4") {
        v4[ip] = mask;
    } else {
        v6[ip] = mask;
    }
}
END {
    for (ip in v4) {
        split(ip, parts, ".");
        prefix = parts[1]"."parts[2];
        mask = v4[ip];
        v4_ranges[prefix] = (prefix in v4_ranges) ? (v4_ranges[prefix]" "mask) : mask;
    }

    for (ip in v6) {
        split(ip, parts, ":");
        prefix = parts[1]":"parts[2]":"parts[3]":"parts[4]":"parts[5]":"parts[6]":"parts[7]":"parts[8];
        mask = v6[ip];
        v6_ranges[prefix] = (prefix in v6_ranges) ? (v6_ranges[prefix]" "mask) : mask;
    }

    for (prefix in v4_ranges) {
        printf "%s/%s\n", prefix, v4_ranges[prefix];
    }

    for (prefix in v6_ranges) {
        printf "%s/%s\n", prefix, v6_ranges[prefix];
    }
}' "$input_file" > "$output_file"

echo "Optimized IP ranges saved to $output_file"
