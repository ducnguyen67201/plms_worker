from typing import List
import sys
import json

class Solution:
	def add(self, a, b, c): return a + b + c

if __name__ == '__main__':
	solution = Solution()
	data = json.loads(sys.stdin.read())
	a = data["a"]
	b = data["b"]
	result = solution.add(a, b)
	print(result)
