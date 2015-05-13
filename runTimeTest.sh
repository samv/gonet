for i in `seq 1 10`; do
	./udp_read_tester &
	python udpWrite.py 2010 $i
	echo $i
done
python timeTest.py Ours
