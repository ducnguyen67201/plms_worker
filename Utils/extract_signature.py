import ast
import sys
import json


def extract_target_method(code):
    tree = ast.parse(code)
    for node in ast.iter_child_nodes(tree):
        if isinstance(node, ast.ClassDef) and node.name == "Solution":
            for subnode in node.body:
                if isinstance(subnode, ast.FunctionDef) and not subnode.name.startswith(
                    "__"
                ):
                    param_names = [
                        arg.arg for arg in subnode.args.args[1:]
                    ]  # Skip 'self'
                    return {"method_name": subnode.name, "param_names": param_names}
    return {}


if __name__ == "__main__":
    code = sys.stdin.read()
    # print(f"Received code:\n{code}", file=sys.stderr)
    result = extract_target_method(code)
    print(json.dumps(result))
