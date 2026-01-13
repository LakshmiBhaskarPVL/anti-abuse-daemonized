rule test_rule {
    meta:
        description = "Test rule to verify YARA scanning works"
        author = "Anti-Abuse"
        severity = "low"
    strings:
        $string1 = "test"
    condition:
        $string1
}

rule empty_file {
    meta:
        description = "Detects empty files"
        author = "Anti-Abuse"
    condition:
        uint32be(0) == 0x00000000
}
