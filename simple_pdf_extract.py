#!/usr/bin/env python3
"""
Simple PDF Text Extraction Script
"""

try:
    from pypdf import PdfReader
except ImportError:
    try:
        from PyPDF2 import PdfReader
    except ImportError:
        print("Installing required PDF library...")
        import subprocess
        import sys
        subprocess.check_call([sys.executable, "-m", "pip", "install", "pypdf"])
        from pypdf import PdfReader

import os

def extract_pdf_text(pdf_path):
    """Extract text from PDF file"""
    print(f"📄 Extracting text from: {pdf_path}")
    
    try:
        with open(pdf_path, 'rb') as file:
            reader = PdfReader(file)
            
            print(f"📃 Total pages: {len(reader.pages)}")
            print("=" * 80)
            
            all_text = []
            
            for page_num, page in enumerate(reader.pages, 1):
                print(f"\n📖 PAGE {page_num}")
                print("-" * 50)
                
                try:
                    text = page.extract_text()
                    if text.strip():
                        print(text)
                        all_text.append(f"--- PAGE {page_num} ---\n{text}")
                    else:
                        print("(No text found on this page)")
                        all_text.append(f"--- PAGE {page_num} ---\n(No text)")
                except Exception as e:
                    error_msg = f"Error extracting page {page_num}: {e}"
                    print(error_msg)
                    all_text.append(f"--- PAGE {page_num} ---\n{error_msg}")
            
            # Save extracted text
            output_file = "pdf_extracted_text.txt"
            with open(output_file, 'w', encoding='utf-8') as f:
                f.write("PDF TEXT EXTRACTION RESULTS\n")
                f.write("=" * 50 + "\n\n")
                f.write("\n\n".join(all_text))
            
            print(f"\n💾 All text saved to: {output_file}")
            
    except Exception as e:
        print(f"❌ Error: {e}")

if __name__ == "__main__":
    pdf_file = "Anti-Bot TLS Fingerprint Task v3.pdf"
    
    if os.path.exists(pdf_file):
        extract_pdf_text(pdf_file)
    else:
        print(f"❌ PDF file not found: {pdf_file}")
        print("📁 Looking for PDF files...")
        for file in os.listdir('.'):
            if file.endswith('.pdf'):
                print(f"   Found: {file}")
                extract_pdf_text(file)
                break
