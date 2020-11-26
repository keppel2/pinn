go run ./src $1 > $1.S && gcc -o $1.out $1.S && ./$1.out || (echo $1 $? && false)
