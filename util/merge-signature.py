# Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


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


if __name__ == "__main__":
    argv = sys.argv
    if len(argv) != 4:
        print("argv is not matched")
        sys.exit(1)

    try:
        merge(*argv[1:])
    except Exception as ex:
        print(ex)
        sys.exit(1)

    sys.exit(0)
