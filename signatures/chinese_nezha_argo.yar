rule CHINESE_NEZHA_ARGO
{
    meta:
        description = "Detects NEZHA/Argo-style scripts (supports both plain and base64 IoCs)"
        author      = "inxtagram (updated)"
        date        = "2025-10-07"
        note        = "All strings have both plain and base64 variants; NEZHA/ARGO_AUTH -> instant match"

    strings:
        $a1  = "nezha"
        $a1_b64  = "nezha" base64
        $a2  = "tunnel.json"
        $a2_b64  = "tunnel.json" base64
        $a3  = "vless"
        $a3_b64  = "vless" base64
        $a4  = "trycloudflare.com"
        $a4_b64  = "trycloudflare.com" base64
        $a5  = "vmess" nocase
        $a5_b64  = "vmess" base64
        $a6  = "WARP"
        $a6_b64  = "WARP" base64
        $a7  = "/eooce/"
        $a7_b64  = "/eooce/" base64
        $a8  = "ARGO_AUTH" nocase
        $a8_b64  = "ARGO_AUTH" base64
        $a9  = "--edge-ip-version"
        $a9_b64  = "--edge-ip-version" base64
        $a10 = "LS1lZGdlLWlwLXZlcnNpb24=" // Base64("--edge-ip-version")
        $a11 = "sub.txt"
        $a11_b64 = "sub.txt" base64
        $a12 = "Server is running on port "
        $a12_b64 = "Server is running on port " base64
        $a13 = "NEZHA" nocase
        $a13_b64 = "NEZHA" base64
        $a14 = "babama1001980"
        $a14_b64 = "babama1001980" base64
        $a15 = "HY2_PORT"
        $a15_b64 = "HY2_PORT" base64
        $a16 = "ssss.nyc.mn"
        $a16_b64 = "ssss.nyc.mn" base64
        $a17 = "using this script,"
        $a17_b64 = "using this script," base64
        $a18 = "Outlook-iOS/696.1102041.prod.iphone (2.99.0)"
        $a18_b64 = "Outlook-iOS/696.1102041.prod.iphone (2.99.0)" base64
        $a19 = "https://speed.cloudflare.com/meta"
        $a19_b64 = "https://speed.cloudflare.com/meta" base64
        $a20 = ">/dev/null"
        $a20_b64 = ">/dev/null" base64
        $a21 = "agent-linux_arm64"
        $a21_b64 = "agent-linux_arm64" base64
        $a22 = "https://1.1.1.1/dns-query"
        $a22_b64 = "https://1.1.1.1/dns-query" base64
        $a23 = "?encryption=none&security=none&host="
        $a23_b64 = "?encryption=none&security=none&host=" base64
        $a24 = "application/dns-message"
        $a24_b64 = "application/dns-message" base64
        $a25 = "--report-delay"
        $a25_b64 = "--report-delay" base64
        $a26 = "--skip-conn"
        $a26_b64 = "--skip-conn" base64
        $a27 = "hysteria2://"
        $a27_b64 = "hysteria2://" base64
        $a28 = "_ = lambda __ : __import__('zlib').decompress"
        $a28_b64 = "_ = lambda __ : __import__('zlib').decompress" base64
        $a29 = "hysteria://"
        $a30 = "hysteria://" base64
        $a31 = "argosbx.sh"
        $a32 = "argosbx.sh" base64
        $a33 = "xtls-rprx-vision"
        $a34 = "ARGO_DOMAIN"
        $a35 = "ARGO_DOMAIN" base64

    condition:
        // Instant hit conditions (plain or base64)
        any of ($a1,$a9,$a1_b64,$a8,$a8_b64,$a12,$a12_b64, $a13,$a13_b64,$a22, $a22_b64, $a27,$a27_b64, $a28, $a28_b64, $a29, $a30, $a31, $a32, $a33, $a34, $a35)
        or
        (
            // grouped IoCs (any form: plain or base64)
            (3 of ($a1,$a1_b64,$a2, $a2_b64,$a3,$a3_b64,$a4,$a4_b64,$a5,$a5_b64,$a6,$a6_b64,$a13,$a15,$a15_b64,$a13_b64,$a18,$a18_b64,$a19,$a19_b64,$a23,$a23_b64))
            and
            (1 of ($a9_b64,$a10,$a11,$a11_b64,$a25,$a25_b64,$a26,$a26_b64,$a7,$a7_b64))
            and
            (1 of ($a24,$a24_b64))
            and
            (1 of ($a14,$a14_b64,$a21,$a21_b64,$a16,$a16_b64,$a17,$a17_b64,$a20,$a20_b64))
        )
}
