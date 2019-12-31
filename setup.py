#!/usr/bin/env python
# -*- encoding: utf-8 -*-

import io
import re

from setuptools import find_packages
from setuptools import setup

with io.open("README.md", "rt", encoding="utf8") as f:
    readme = f.read()

with io.open("src/__init__.py", "rt", encoding="utf8") as f:
    version = re.search(r'__version__ = "(.*?)"', f.read()).group(1)

setup(
        name="nsddyn",
        version=version,
        url="https://gitlab.com/necheffa/nsddyn",
        project_urls={
            "Documentation": "https://gitlab.com/necheffa/nsddyn",
            "Code": "https://gitlab.com/necheffa/nsddyn",
            "Issue tracker": "https://gitlab.com/necheffa/nsddyn",
        },
        license="GPLv3",
        author="Alexander Necheff",
        author_email="alex@necheff.net",
        maintainer="Alexander Necheff",
        maintainer_email="alex@necheff.net",
        description="nsddyn provides a secure method for achieving Dynamic DNS when using NSD as an authoritative DNS server.",
        long_description=readme,
        classifiers=[
            "Development Status :: 1 Planning",
            "Intended Audience :: System Administrators",
            "License :: OSI Approved :: GNU General Public License v3 (GPLv3)",
            "Operating System :: POSIX",
            "Operating System :: Unix",
            "Programming Language :: Python",
            "Programming Language :: Python :: 3",
            "Programming Language :: Python :: Implementation :: CPython",
            "Programming Language :: Python :: Implementation :: PyPy",
            "Topic :: Internet :: Name Service (DNS)",
        ],
        packages=find_packages("src"),
        package_dir={"": "src"},
        include_package_data=True,
        python_requires=">=3.7",
        install_requires=[
            "wheel>=0.33.6",
            "localzone>=0.9.5",
        ],
        extras_require={
            "dev": [
                "pytest",
                "coverage"
            ],
        },
    )

