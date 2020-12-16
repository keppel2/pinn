./gr.sh $* `uname -m` > $1.S && ./da.sh $1 || (echo :$1 $? > /dev/stderr && false)
