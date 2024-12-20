# check if all files are formatted
check-format:
	@OUTPUT=$$(gofmt -l .); \
	if [ -n "$$OUTPUT" ]; then \
		echo "The following files aren't formatted. Please run 'make format':"; \
		echo "$$OUTPUT"; \
		exit 1; \
	fi

# format all files
format:
	gofmt -w .

