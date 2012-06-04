import os
from setuptools import setup, find_packages

# Utility function to read the README file.
# Used for the long_description.  It's nice, because now 1) we have a top level
# README file and 2) it's easier to type in the README file than to put a raw
# string in below ...
def read(fname):
        f = open(os.path.join(os.path.dirname(__file__), fname))
        long_desc = f.read()
        f.close()
        return long_desc

VERSION = __import__('hjoin').__version__

setup(
    name="hatcog",
    version=VERSION,
    author='Graham King',
    author_email='graham@gkgk.org',
    description="IRC client, text based, perfect for tmux or screen",
    long_description=read('README.md'),
    packages=find_packages(),
    data_files=[('/usr/local/bin', ['bin/hatcogd-32', 'bin/hatcogd-64'])],
    entry_points={
        'console_scripts':[
            'hjoin=hjoin.hjoin:main'
            ]
    },
    url="https://github.com/grahamking/hatcog",
    install_requires=['setuptools'],
    classifiers=[
        'Development Status :: 4 - Beta',
        'Environment :: Console',
        'Environment :: Console :: Curses',
        'Intended Audience :: End Users/Desktop',
        'License :: OSI Approved :: GNU General Public License (GPL)',
        'Operating System :: POSIX',
        'Programming Language :: Python :: 3',
        'Programming Language :: Go',
        'Topic :: Communications :: Chat :: Internet Relay Chat',
    ]
)
