./gr.sh $* `uname -m` > $1.S && ./da.sh $1 || ./er.sh $1 $?
