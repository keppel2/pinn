./gr.sh $1 `uname -m` > $1.S && ./da.sh $1 || (echo $1 $? && false)
