import sys
import difflib
import re
from collections import defaultdict
from pypdf import PdfReader


def extract_text_from_pdf(pdf_path):
    """从PDF文件中提取文本"""
    reader = PdfReader(pdf_path)
    text = ""
    for page in reader.pages:
        text += page.extract_text() + "\n"

    return text

def extract_text(pdf_path):
    """从PDF文件中提取文本"""
    # Reading the file
    with open(pdf_path, "r") as file:
        return "\n".join(file.readlines())




def split_into_words(text):
    """将文本分割为单词列表（保留空格和标点）"""
    return re.findall(r'\w+|\s+|\W+', text)

def generate_word_level_diff(text1, text2):
    """生成单词级别的行内差异标记"""
    words1 = split_into_words(text1)
    words2 = split_into_words(text2)

    differ = difflib.SequenceMatcher(None, words1, words2)
    result = []

    for tag, i1, i2, j1, j2 in differ.get_opcodes():
        if tag == 'equal':
            result.extend(words1[i1:i2])
        elif tag == 'delete':
            result.append(f"[-{' '.join(words1[i1:i2])}-]")
        elif tag == 'insert':
            result.append(f"{{+{' '.join(words2[j1:j2])}+}}")
        elif tag == 'replace':
            result.append(f"[-{' '.join(words1[i1:i2])}-]{{+{' '.join(words2[j1:j2])}+}}")

    return ''.join(result)

def compare_texts(text1, text2):
    """比较两个文本并生成行内差异报告"""
    lines1 = text1.splitlines()
    lines2 = text2.splitlines()

    differ = difflib.Differ()
    diff = list(differ.compare(lines1, lines2))

    result = []
    buffer_removed = []
    buffer_added = []

    for line in diff:
        if line.startswith('  '):  # 未改变的行
            if buffer_removed or buffer_added:
                result.append(generate_word_level_diff(
                    '\n'.join(buffer_removed),
                    '\n'.join(buffer_added)
                ))
                buffer_removed = []
                buffer_added = []
            result.append(line[2:])
        elif line.startswith('- '):  # 仅在文档1中的行
            buffer_removed.append(line[2:])
        elif line.startswith('+ '):  # 仅在文档2中的行
            buffer_added.append(line[2:])

    # 处理剩余的缓冲区内容
    if buffer_removed or buffer_added:
        result.append(generate_word_level_diff(
            '\n'.join(buffer_removed),
            '\n'.join(buffer_added)
        ))

    return '\n'.join(result)

def save_diff_report1(diff_text, output_path):
    """保存差异报告到文件"""
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(diff_text)

def save_diff_report(diff_text, output_path):
    """保存差异报告到HTML文件"""
    html_content = f"""
        <pre>{diff_text.replace('[-', '<span class="removed">').replace('-]', '</span>')
                     .replace('{+', '<span class="added">').replace('+}', '</span>')}</pre>
    """
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(html_content)



def compare_pdfs(pdf1_path, pdf2_path, output_path):
    """主函数：比较两个PDF并生成差异报告"""
    print(f"正在提取 {pdf1_path} 的文本...")
    text1 = extract_text_from_pdf(pdf1_path)

    print(f"正在提取 {pdf2_path} 的文本...")
    text2 = extract_text_from_pdf(pdf2_path)

    print("正在生成单词级差异报告...")
    diff_text = compare_texts(text1, text2)

    print(f"保存报告到 {output_path}...")
    save_diff_report(diff_text, output_path)

    print("比较完成！")

if __name__ == "__main__":
    argv = sys.argv
    if len(argv) != 4:
        print("argv is not matched")
        sys.exit(1)
    # 配置输入文件和输出文件路径
    PDF_FILE_1 = argv[1]  # 第一个PDF文件路径
    PDF_FILE_2 = argv[2]  # 第二个PDF文件路径
    OUTPUT_FILE = argv[3]  # 输出文件路径

    # 执行比较
    compare_pdfs(PDF_FILE_1, PDF_FILE_2, OUTPUT_FILE)
