package main

import "fmt"

func main() {
	fmt.Printf("=== TestChangeSet_Merge 问题分析 ===\n\n")
	
	fmt.Printf("原始文档: \"hello\" (长度 5)\n\n")
	
	fmt.Printf("cs1 = NewChangeSet(5).Retain(5).Insert(\" world\")\n")
	fmt.Printf("  - lenBefore = 5\n")
	fmt.Printf("  - Retain(5): 跳过 \"hello\"\n")
	fmt.Printf("  - Insert(\" world\", 6 chars): 在位置 5 插入\n")
	fmt.Printf("  - lenAfter = 5 + 6 = 11\n\n")
	
	fmt.Printf("cs2 = NewChangeSet(5).Retain(11).Insert(\"!\")\n")
	fmt.Printf("  - lenBefore = 5\n")
	fmt.Printf("  - Retain(11): 试图跳过 11 个字符\n")
	fmt.Printf("  - 但是文档只有 5 个字符！\n")
	fmt.Printf("  - 这是一个 INVALID 的 changeset！\n\n")
	
	fmt.Printf("测试用例期望合并这两个 changesets 得到 \"hello world!\"，\n")
	fmt.Printf("但 cs2 本身就无法应用到原始文档上。\n\n")
	
	fmt.Printf("正确的做法应该是:\n")
	fmt.Printf("cs2 := NewChangeSet(cs1.LenAfter()).Retain(11).Insert(\"!\")\n")
	fmt.Printf("  - 这样 cs2 就是基于 cs1 的结果创建的\n")
	fmt.Printf("  - 应该使用 Compose 而不是 Merge\n")
}
