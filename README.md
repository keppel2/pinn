# Pinn to llvm/ARM64 project. PinnProject.

Implement Pinn entirely in golang (remove ANTLR), outputting LLVM initially, and ARM64 later.

## Motivations

* Repeatedly saw interpreted languages (Ruby, Python), slower by orders of magnitude than Go.
* ANTLR works well, but a clean hand written implementation should be faster. Also could not be there for a self-hosting eventual solution.
* LLVM works well for clang and Rust. Probably can't reach the speed of native--can check Gollvm vs Golang(self-hosted).
* ARM64 is set to appear on Macs. Owns phone space. Windows 10 now runs on it. It's fairly clean (especially compared to x64).

## x64

* Game consoles and most Windows. Switch runs ARM.

## Steps

* Lexer. `text/scanner` looks like a good library.
* Parser. Follow example in `src/cmd/compile/internal/syntax`--used for self-hosted Go compilation and self-contained.
* Generate LLVM IR. https://github.com/llir/llvm is a Go library that can be used to generate the IR.
