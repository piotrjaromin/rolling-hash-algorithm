# Rolling hash algorithm

Simple implementation of file rolling hash algorithm based on rsync idea.
Currently supports creating signature file and delta files.
For simplicity buffer and chunk size are set to small values, but in real live those should be way bigger.

https://www.andrew.cmu.edu/course/15-749/READINGS/required/cas/tridgell96.pdf

## Usage

(first build with `make build`)

examples

```bash
./bin/sync signature --inputFile testfile.txt --signatureFile sig.txt
```

```bash
./bin/sync delta --inputFile testfile.txt --signatureFile sig.txt --deltaFile delta.txt
```