from setuptools import setup
import os

def read_requirements():
    with open('requirements.txt', 'r') as f:
        return [line.strip() for line in f if line.strip() and not line.startswith('#')]

setup(
    name='gtasks2md',
    version='0.1.0',
    description='Google Tasks to Markdown Sync',
    py_modules=['auth', 'cli', 'core', 'google_tasks', 'main', 'markdown', 'models'],
    install_requires=read_requirements(),
    entry_points={
        'console_scripts': [
            'google-tasks=cli:cli',
        ],
    },
)
