#!/usr/bin/env python3
"""
Setup script for Engine Language.
"""

from setuptools import setup, find_packages
import os

# Read the README file for long description
def read_readme():
    readme_path = os.path.join(os.path.dirname(__file__), 'README.md')
    if os.path.exists(readme_path):
        with open(readme_path, 'r', encoding='utf-8') as f:
            return f.read()
    return "Engine Language - A fast and simple interpreted programming language"

setup(
    name="engine-lang",
    version="0.1.0",
    author="Engine Language Team",
    author_email="",
    description="A fast and simple interpreted programming language",
    long_description=read_readme(),
    long_description_content_type="text/markdown",
    url="https://github.com/yourusername/Engine-Lang",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Software Development :: Interpreters",
        "Topic :: Software Development :: Compilers",
    ],
    python_requires=">=3.7",
    entry_points={
        "console_scripts": [
            "engine=engine.__main__:main",
        ],
    },
    include_package_data=True,
    zip_safe=False,
)
