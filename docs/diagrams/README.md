# Architecture Diagram Generator

This project generates an architectural diagram using the `Diagrams` Python package.

## Prerequisites

- Install and configure `pyenv` to manage Python versions:
  1. Install `pyenv` by following the instructions at [pyenv GitHub](https://github.com/pyenv/pyenv#installation).
  2. Install Python 3.12.3 or higher with `pyenv`:
     ```bash
     pyenv install 3.12.3 
     ```
  3. Set the local Python version for this project:
     ```bash
     pyenv local 3.12.3 
     ```
- Ensure `pip` is available for the selected Python version:
     ```bash
     python3 -m ensurepip --upgrade
     ```

## Setup
1. Install the Graphviz system package
   ```bash
   sudo apt install graphviz
   ```

2. Install the required dependencies in a virtual environment:
   ```bash
   python3 -m venv venv  # Create a virtual environment
   source venv/bin/activate  # Activate the virtual environment
   pip install -r requirements.txt  # Install dependencies
   ```

3. Verify the environment is set up correctly:
   ```bash
   python3 --version
   ```

## Running the Script

1. Activate the virtual environment if not already active:
   ```bash
   source venv/bin/activate
   ```

2. Run the script to generate the architecture diagram:
   ```bash
   python3 gen-arch.py
   ```

3. The diagram will be saved as `System Architecture.png` in the current directory.

## Files

- `gen-arch.py`: Python script to generate the diagram.
- `requirements.txt`: Contains the required Python packages.
- `README.md`: This documentation.

## Notes

- To deactivate the virtual environment after use, run:
   ```bash
   deactivate
   ```
- To switch between Python versions using `pyenv`, use:
   ```bash
   pyenv local <version>
   ```
