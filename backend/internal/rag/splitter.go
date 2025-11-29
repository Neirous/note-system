package rag

import (
    "regexp"
    "strings"
)

type FragCandidate struct {
    Content string
    IsCode  bool
}

func SplitMarkdown(md string) []FragCandidate {
    res := make([]FragCandidate, 0)
    codeBlockRe := regexp.MustCompile("(?s)```.*?```")
    idxs := codeBlockRe.FindAllStringIndex(md, -1)
    last := 0
    for _, rng := range idxs {
        if rng[0] > last {
            plain := strings.TrimSpace(md[last:rng[0]])
            addParagraphs(&res, plain)
        }
        code := strings.TrimSpace(md[rng[0]:rng[1]])
        if code != "" { res = append(res, FragCandidate{Content: code, IsCode: true}) }
        last = rng[1]
    }
    if last < len(md) {
        plain := strings.TrimSpace(md[last:])
        addParagraphs(&res, plain)
    }
    return res
}

func addParagraphs(out *[]FragCandidate, text string) {
    if text == "" { return }
    parts := splitByBlank(text)
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" { continue }
        for len(p) > 0 {
            chunk := p
            if len(chunk) > 500 {
                chunk = chunk[:500]
            }
            *out = append(*out, FragCandidate{Content: chunk, IsCode: false})
            if len(p) <= 500 { break }
            p = p[500:]
        }
    }
}

func splitByBlank(text string) []string {
    re := regexp.MustCompile("\n\n+")
    return re.Split(text, -1)
}

