go run ./src $1 rp > $1.S && gcc -o $1.out $1.S && ./$1.out
