all: book.html book.pdf
	@echo DONE

book.html book.pdf: book.md
	pandoc -o book.html -f markdown_github --toc -t html -s book.md
	pandoc -o book.pdf -f markdown_github --toc -t latex book.md

