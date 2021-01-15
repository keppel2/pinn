package main
import "fmt"
const (
 SIZER = 1;
 SIZEC = 5;
 BSIZE =  SIZER * SIZEC


E = 0;
B = 1;
W = 2;
N = 3;
X = 4;
I = 5;
TRUE = 1;
FALSE = 0;

)

var HASHT =  pow(3, BSIZE)
var  BTOT = HASHT * BSIZE + BSIZE

func pow (a, b int) int {
  rt := 1
  for x := 0; x < b; x++ {
    rt *= a
  }
  return rt
}


var _counter int;
var bd = make([]int, BTOT);
var bh = make([]int, HASHT);

var gohash = make(map[int]bool)

func move (player, r, c int) int {
  push();
  bd[hrc(hsize - 1, r, c)] = player;
  cap(opposite(player), r + 1, c);
  cap(opposite(player), r - 1, c);
  cap(opposite(player), r, c + 1);
  cap(opposite(player), r, c - 1);
  cap(player, r, c);
  return sko();
}

func hash() int {
  rt := 0;
  var r, c, hc int;
  for r = 0; r < SIZER; r++ {
    for c = 0; c < SIZEC; c++ {
      rt = rt * 3;
      hc = hrc(hsize - 1, r, c); 
      rt = rt + bd[hc];
    }
  }
  return rt;

}
func sko() int {
  h := hash();
  if bh[h] == 1 {
    hsize--;
    return FALSE;
  }
  bh[h] = 1;
  return TRUE;
}

func mark (player, r, c int) int {
  var rt int;
  if r < 0 || r >= SIZER || c < 0 || c >= SIZEC || cp[rc(r, c)] == opposite(player) || cp[rc(r, c)] == I {
    return FALSE;
  }
  if cp[rc(r, c)] == E {
    return TRUE;
  }
  cp[rc(r, c)] = I;
  rt = FALSE;
  if mark(player, r + 1, c) == TRUE { rt = TRUE; }
  if mark(player, r - 1, c) == TRUE { rt = TRUE; }
  if mark(player, r, c + 1) == TRUE { rt = TRUE; }
  if mark(player, r, c - 1) == TRUE { rt = TRUE; }
  return rt;
}

func cap (player, r, c int) {

  if r < 0 || r >= SIZER || c < 0 || c >= SIZEC {
    return;
  }
  var tr, tc int;
  copy2cp();
  rt := mark(player, r, c);
  if rt == TRUE {
    return;
  }
  for tr = 0; tr < SIZER; tr++ {
    for tc = 0; tc < SIZEC; tc++ {
      if cp[rc(tr, tc)] == I {
        cp[rc(tr, tc)] = E;
      }
    }
  }
  copy2bd();
}
var cp [BSIZE] int;
func copy2bd() {
  var r int;
  var c int;
  for r = 0; r < SIZER; r++ {
    for c = 0; c < SIZEC; c++ {
      bd[hrc(hsize - 1, r, c)] = cp[rc(r, c)];
    }
  }
}

func push() {
  hsize++;
  var r int;
  var c int;
  for r = 0; r < SIZER; r++ {
    for c = 0; c < SIZEC; c++ {
      bd[hrc(hsize - 1, r, c)] = bd[hrc(hsize - 2, r, c)];
    }
  }
}

func copy2cp() {
  var r int;
  var c int;
  var dd int;
  for r = 0; r < SIZER; r++ {
    for c = 0; c < SIZEC; c++ {
      dd = bd[hrc(hsize - 1, r, c)];
      cp[rc(r, c)] = dd;

    }
  }
}

func score (h int) int {
    b := 0;
    w := 0;
    var cs int;
    copy2cp();
    var r, c, p int;
    for r = 0; r < SIZER; r++ {
        for c = 0; c < SIZEC; c++ {
            if bd[hrc(h, r, c)] == B {b++;} else
            if bd[hrc(h, r, c)] == W {w++;} else
            {
              p = N;
                  cs, p = color(p, r, c, 0);
                  if p == B {b += cs;}
                  if p == W {w += cs;}
            }
            
        }
    }
    return b - w;
}

func mscore(player int) int {
  if player == B {
    return score(hsize -1 )
  }
  return -score(hsize -1 )
}

func color (p, r, c, score int) (int, int) {
    if r < 0 || r >= SIZER || c < 0 || c >= SIZEC || cp[rc(r, c)] == I {
      return score, p;
    }
    pin := cp[rc(r, c)];
    if pin == E {
      cp[rc(r,c)] = I;
      score++;
      score, p = color(p, r + 1, c, score);
      score, p = color(p, r - 1, c, score);
      score, p = color(p, r, c + 1, score);
      score, p = color(p, r, c - 1, score);
    } else if p == N {
      p = pin;
    } else if p != pin {
      p = X;
    }
    return score, p;
}

func rc (r, c int) int {
  return r * SIZEC  + c;
}
func hrc (h, r, c int) int {
  return  h * BSIZE + rc(r, c);
}

func opposite (opx int) int { 
  if opx == B {
    return W
  }
  return B
}

var hsize = 1;

func main() {
  bh[hash()] = 1;
  hsize++
  bh[hrc(hsize - 1, 0, 2)] = B
  bh[hash()] = 1
  hsize++
  sc := minimax(W, FALSE);
  fmt.Println(sc);
  fmt.Println(_counter);
}


func minimax (player int, passed int) int {
  if _counter % 0x10_000_000 == 0 {
    fmt.Print(_counter)
    fmt.Print(",")
  }
  _counter++;
  var best int;
  if passed == TRUE {
    best = mscore(player);
  } else {
    best = -minimax(opposite(player), TRUE);
  }
  if best == BSIZE {
    return best;
  }

  var s, r, c, b  int;
  for r = 0; r < SIZER; r++ {
    for c = 0; c < SIZEC; c++ {
      if bd[hrc(hsize - 1, r, c)] == E {
        b = move(player, r, c);
        if b == TRUE {
          s = -minimax(opposite(player), FALSE);
          if s > best {
            best = s;
          }
          bh[hash()] = 0;
          hsize--;
          if best == BSIZE {
            return best;
          }
        }
      }
    }
  }
  return best;
}

