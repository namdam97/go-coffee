buf.gen:
	docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf generate
echo:
	echo $(PWD)