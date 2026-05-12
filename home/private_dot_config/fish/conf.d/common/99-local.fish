# Machine-specific overrides — not versioned, managed directly on each machine
# Fish equivalent of ~/.env.local (bash syntax won't work here)
if test -f ~/.env.local.fish
    source ~/.env.local.fish
end
