go run ./src $1 > $1.s && gcc -o $1.out $1.s && ./$1.out
