import sys

from PyPDF2 import PdfFileReader
from PyPDF2 import PdfFileWriter
from pathlib import Path


def merge(pdf_file, org_signature, out_file):
    writer = PdfFileWriter()

    pdf = PdfFileReader(pdf_file)
    num = pdf.getNumPages()
    for i in range(num - 1):
        writer.addPage(pdf.getPage(i))

    pdf1 = PdfFileReader(org_signature)
    page = pdf1.getPage(0)
    page.mergePage(pdf.getPage(num - 1))
    writer.addPage(page)
    
    with Path(out_file).open("wb") as out:
        writer.write(out)


def append(pdf_file, org_signature, out_file):
    writer = PdfFileWriter()

    pdf = PdfFileReader(pdf_file)
    writer.appendPagesFromReader(pdf)

    pdf1 = PdfFileReader(org_signature)
    page = pdf1.getPage(0)
    writer.addPage(page)

    with Path(out_file).open("wb") as out:
        writer.write(out)


if __name__ == "__main__":
    argv = sys.argv
    if len(argv) != 5:
        print("argv is not matched")
        sys.exit(1)

    f = merge
    if argv[1] == "append":
        f = append

    try:
        f(*argv[2:])
    except Exception as ex:
        print(ex)
        sys.exit(1)

    sys.exit(0)
