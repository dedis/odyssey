#!/bin/sh

test_folder="/tmp/cryptutil"
cryptutil=$test_folder/cryptutil

main() {
    setUp
    testEncrypt
    testDecrypt
    endUp
}

# ------------------------------------------------------------------------------
# START Utilities
setUp() {
    set -e
    echo "* Starting test..."
    mkdir -p $test_folder
    go build -o $cryptutil
}

fail(){
  echo
  echo -e "\tFAILED: $@"
  endUp
  exit 1
}

endUp() {
    echo "* ...ending test."
    # rm -rf $test_folder
}

testOK(){
  echo "Assert OK for '$@'"
  if ! "$@"; then
    fail "starting $@ failed"
  fi
}

testFail(){
  echo "Assert FAIL for '$@'"
  if "$@"; then
    fail "starting $@ should've failed, but succeeded"
  fi
}

matchOK(){
  echo "Match OK between '$1' and '$2'"
  if ! [[ $1 =~ $2 ]]; then
    fail "'$1' does not match '$2'"
  fi
}
# END utilities
# ------------------------------------------------------------------------------

testEncrypt() {
    echo "* testEncrypt"
    testFail $cryptutil encrypt
    testFail $cryptutil encrypt --data "Hello world."
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeeff"
    testOK $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb"
    # wrong key length / type
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeef" --initVal "00112233445566778899aabb"
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddee" --initVal "00112233445566778899aabb"
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeefg" --initVal "00112233445566778899aabb"
    # wrong init val length / type
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aab"
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aa"
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabg"
    # wrong key and init valu
    testFail $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeef" --initVal "00112233445566778899aab"
    testFail $cryptutil encrypt --data "Hello world." --key "12" --initVal "abcd"
    testFail $cryptutil encrypt --data "Hello world." --key "xasf3sf" --initVal "2132ga"
    # good keyAndInitVal
    testOK $cryptutil encrypt  --data "Hello world." --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    testOK $cryptutil encrypt  --data "Hello world." --key "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    testOK $cryptutil encrypt  --data "Hello world." --initVal "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    testOK $cryptutil encrypt  --data "Hello world." --key "aaa" --initVal "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    # wrong keyAndInitVal
    testFail $cryptutil encrypt  --data "Hello world." --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aab"
    testFail $cryptutil encrypt  --data "Hello world." --key "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabbb"
    testFail $cryptutil encrypt  --data "Hello world." --initVal "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabx"
    testFail $cryptutil encrypt  --data "Hello world." --key "aaa" --initVal "aaa" --keyAndInitVal "a"

    # Check output
    OUTRES=$($cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb")
    expected="ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09"
    matchOK "$OUTRES" "^$expected$"
    # Check output with keyAndInitVal
    OUTRES=$($cryptutil encrypt --data "Hello world." --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb")
    matchOK "$OUTRES" "^$expected$"
    # Now providind both --key / --initVal and --keyAndInitVal
    # --keyAndInitVal should be used
    OUTRES=$($cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeefa" --initVal "00112233445566778899aaba" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb")
    matchOK "$OUTRES" "^$expected$"

    # Data from file
    printf 'Hello world.'  > $test_folder/hello_world.txt
    OUTRES=$($cryptutil encrypt --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb" --readData < $test_folder/hello_world.txt)
    matchOK "$OUTRES" "^$expected$"

    # Output to stdout. As this saves raw encrypted data and not hex encoded, we
    # will assert the content we the decrypt function.
    testOK $cryptutil encrypt --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb" --readData -x < $test_folder/hello_world.txt > $test_folder/hello_world.txt.aes
}

testDecrypt() {
    echo "* testDecrypt"
    testFail $cryptutil decrypt
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeeff"
    testOK $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb"
    # wrong key length / type
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeef" --initVal "00112233445566778899aabb"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddee" --initVal "00112233445566778899aabb"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeefg" --initVal "00112233445566778899aabb"
    # wrong init val length / type
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aab"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aa"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabg"
    # wrong key and init val
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeef" --initVal "00112233445566778899aab"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "12" --initVal "abcd"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "xasf3sf" --initVal "2132ga"
    # good keyAndInitVal
    testOK $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    testOK $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    testOK $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --initVal "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    testOK $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "aaa" --initVal "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb"
    # wrong keyAndInitVal
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aab"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabbb"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --initVal "aaa" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabx"
    testFail $cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "aaa" --initVal "aaa" --keyAndInitVal "a"

    # Check output
    OUTRES=$($cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb")
    expected="Hello world."
    matchOK "$OUTRES" "^$expected$"
    # Check output with keyAndInitVal
    OUTRES=$($cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb")
    matchOK "$OUTRES" "^$expected$"
    # Now providind both --key / --initVal and --keyAndInitVal
    # --keyAndInitVal should be used
    OUTRES=$($cryptutil decrypt --data "ef5a516eddc4a6656a3f17351b7ffebbe9f8b8c8b282470b59b72c09" --key "00112233445566778899aabbccddeefa" --initVal "00112233445566778899aaba" --keyAndInitVal "00112233445566778899aabbccddeeff00112233445566778899aabb")
    matchOK "$OUTRES" "^$expected$"

    # Data from file 
    # First we must use "encrypt" in order to creat the encrypted file
    # containing raw data.
    $cryptutil encrypt --data "Hello world." --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb" -x > $test_folder/hello_world.txt.aes
    OUTRES=$($cryptutil decrypt --key "00112233445566778899aabbccddeeff" --initVal "00112233445566778899aabb" --readData < $test_folder/hello_world.txt.aes)
    matchOK "$OUTRES" "^$expected$"
}

main