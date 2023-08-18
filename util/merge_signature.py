import sys

from pypdf import PdfReader
from pypdf import PdfWriter
from pathlib import Path


def merge(pdf_file, org_signature, out_file):
    writer = PdfWriter()

    pdf = PdfReader(pdf_file)
    num = pdf.getNumPages()
    for i in range(num - 1):
        writer.add_page(pdf.pages[i])

    pdf1 = PdfReader(org_signature)
    page = pdf1.pages[0]
    page.mergePage(pdf.pages[num - 1])
    writer.add_page(page)

    with Path(out_file).open("wb") as out:
        writer.write(out)


def append(pdf_file, org_signature, out_file):
    writer = PdfWriter()

    pdf = PdfReader(pdf_file)
    writer.append_pages_from_reader(pdf)

    pdf1 = PdfReader(org_signature)
    page = pdf1.pages[0]
    writer.add_page(page)

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
