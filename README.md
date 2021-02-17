# Pinn.

Implement Pinn entirely in golang (remove ANTLR) in x64/ARM64 assembly (eventually binaries).

## Pinn
- Initial implementation in Go/ANTLR, https://github.com/keppel2/pinn-go. Further development in Swift/ANTLR, https://github.com/keppel2/pinn. Further development in this repo.

## Motivations

- Repeatedly saw interpreted languages (Ruby, Python), slower by orders of magnitude than Go.
- ANTLR works well, but a clean hand written implementation should be faster. Also could not be there for a self-hosting eventual solution.
- ARM64 is set to appear on Macs. Owns phone space. Windows 10 now runs on it. It's fairly clean (especially compared to x64).

## Steps

- Lexer. `text/scanner` mostly works because of similarity to Go.
- Parser. LL(1). Small hack to lower precedence of range `:` operator in ternary `? :` expressions.
- Generate x64/ARM64 assembly. ARM64 is lapsed. Mac and Linux are supported but latest is Mac/x64.
  - Output binaries for Mac/Linux eventually.

## Tic-tac-toe solver as a unit test.

```
#define SIZE 3;
#define SQSIZE SIZE SIZE *;

//TIC TAC TOE as unit test
EMPTY := 0;
FALSE := 0;
TRUE := 1;
TIE := 3;
PLAYER_A := 1;
PLAYER_B := 2;
count := 0;

func minimax (player int, depth int, flob [SQSIZE]int) int
{
 var da, db, x, y int;
  var mxy int;
 count++;
 i := 0;
 result := 0;
 var best int;
da = dbg();
  best = opposite(player);
  result = winner(flob);
        db = dbg();
        assert(da, db);

  if result != TIE {
    return result;
  }
  if full(flob) == TRUE {
        return TIE;
  }
  x = 0;
  y = 0;
  for x = 0; x < SIZE; x++ {
    for y = 0; y < SIZE; y++ {

      if flob[xy(x, y)]  == EMPTY {
        flob[xy(x, y)] = player;
        result = minimax(opposite(player), depth + 1, flob);
        if result == player {
          return player;
        }
        if result == TIE {
          best = TIE;
        }
        flob[xy(x, y)] = EMPTY;
 
      }

    }
  }
  return best;
}

func printBoard (flob [SQSIZE]int) {
  println();
  var i, j int;
  for i = 0; i < SIZE; i++ {
    for j = 0; j < SIZE; j++ {
      print(flob[xy(i, j)]);
    }
    println();
  }
  println();
}

func xy (fx, fy int) int {
  rt := fy * SIZE;
  rt += fx;
  return rt;
}


func full (flob [SQSIZE]int) int {
  i := 0;
  for i = 0; i < SQSIZE; i++ {
    if flob[i] == EMPTY {
      return FALSE;
    }
  }
  return TRUE;
}

func line (x int, y int, dx int, dy int, flob [SQSIZE]int) int {
    comp := flob[xy(x, y)];
    if comp == EMPTY {
        return TIE;
    }
    loop {
      if x + dx < SIZE {
        if y + dy < SIZE {
          x += dx;
          y += dy;
          if flob[xy(x, y)] != comp {
            return TIE;
          }
          continue;
        }
      }
      break;
    }
    return comp;
}

func winner (flob [SQSIZE]int) int {
  var ii int;
  rt := 0;
  for ii = 0; ii < SIZE; ii++ {

        rt = line(ii, 0, 0, 1, flob);
        if rt != TIE
            return rt;
        rt = line(0, ii, 1, 0, flob);
        if rt != TIE
            return rt;
    }
    rt = line(0, 0, 1, 1, flob);
    if rt != TIE {
        return rt;
    }
    rt = line(0, SIZE - 1, 1, -1, flob);
    return rt;
}

func opposite (opx int) int { 
 if opx == PLAYER_A {
   return PLAYER_B;
 }
 return PLAYER_A;
 }


func main() {
  var flob [SQSIZE]int;
  x := minimax(PLAYER_A, 0, flob);
  assert(x, TIE);
  assert(count, 0x17302);
  
}
main();
```
