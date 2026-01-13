rule simpletestrule
{
    meta:
        description = "Simple test rule"
    strings:
        $test = "hello"
    condition:
        $test
}
