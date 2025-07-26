#!/bin/bash

# Enhanced Anti-Bot Scraper Demo Script
# This script demonstrates the new configurable inputs and enhanced detection evasion features

echo "=== Anti-Bot Scraper Enhanced Features Demo ==="

# 1. Single URL with enhanced header mimicry
echo "1. Testing enhanced header mimicry with Chrome profile..."
./scraper -url https://httpbin.org/headers \
          -header-mimicry=true \
          -header-profile=chrome \
          -enable-sec-headers=true \
          -verbose

echo -e "\n" && sleep 2

# 2. Multiple URLs from file with TLS randomization
echo "2. Testing multiple URLs with randomized TLS profiles..."
./scraper -urls-file examples/urls.txt \
          -num-requests=2 \
          -tls-randomize=true \
          -delay-min=500ms \
          -delay-max=2s \
          -delray-randomize=true \
          -verbose

echo -e "\n" && sleep 2

# 3. Cookie persistence and redirect handling
echo "3. Testing enhanced cookie and redirect handling..."
./scraper -url https://httpbin.org/redirect/3 \
          -cookie-jar=true \
          -cookie-persistence=session \
          -follow-redirects=true \
          -max-redirects=5 \
          -verbose

echo -e "\n" && sleep 2

# 4. Multiple requests with proxy rotation (comment out if no proxies)
echo "4. Testing proxy rotation with file-based proxy list..."
# ./scraper -url https://httpbin.org/ip \
#           -proxy-file examples/proxies.txt \
#           -num-requests=5 \
#           -verbose

echo -e "\n" && sleep 2

# 5. Advanced evasion with all features enabled
echo "5. Testing comprehensive evasion features..."
./scraper -url https://httpbin.org/headers \
          -header-mimicry=true \
          -header-profile=auto \
          -tls-profile=firefox \
          -custom-user-agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:91.0) Gecko/20100101 Firefox/91.0" \
          -cookie-jar=true \
          -follow-redirects=true \
          -enable-sec-headers=false \
          -accept-language="en-US,en;q=0.5" \
          -accept-encoding="gzip, deflate, br" \
          -output=json \
          -verbose

echo -e "\n=== Demo Complete ===\n"
