{
	n++
	if (n <= 120) {
		print
		fflush()
		system("sleep 0.015")
	}
}

END {
	printf "%d residential proxy events captured in 5s\n", n
}
