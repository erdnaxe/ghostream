from setuptools import setup, find_packages

setup(
    name='ghostream',
    version='0.1.0',
    packages=find_packages(),
    include_package_data=True,
    install_requires=[
        'flask>=1.0.2',
        'python-ldap>=3.1.0',
    ],
)
