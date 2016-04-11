puts RUBY_VERSION

def fib(n)
  (n == 0 || n == 1) ? 1 : fib(n - 1) + fib(n - 2)
end

# never mind - it will never finish
(1..10000).reduce([]) { |acc, n| acc + [fib(n)]}
