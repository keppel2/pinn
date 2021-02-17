# Pinn.

Implement Pinn entirely in golang (remove ANTLR) in x64/ARM64 assembly (eventually binaries).

## Pinn
* Initial implementation in Go/ANTLR, https://github.com/keppel2/pinn-go. Further development in Swift/ANTLR, https://github.com/keppel2/pinn.

## Motivations

* Repeatedly saw interpreted languages (Ruby, Python), slower by orders of magnitude than Go.
* ANTLR works well, but a clean hand written implementation should be faster. Also could not be there for a self-hosting eventual solution.
* ARM64 is set to appear on Macs. Owns phone space. Windows 10 now runs on it. It's fairly clean (especially compared to x64).

## x64

* Game consoles and most Windows. Switch runs ARM.

## Steps

* Lexer. `text/scanner` mostly works because of similarity to Go.
* Parser. LL(1). Small hack to lower precedence of range `:` operator in ternary `? :` expressions.
* Generate x64/ARM64 assembly natively. ARM64 is lapsed. Mac and Linux are supported but latest is Mac/x64.
