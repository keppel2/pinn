./pp.sh pparse && for i in {1..5}; do ./mt.sh em$i || exit 1 ; done  && ./pr.sh pr1 && rm *.S *.out pr1.txt
