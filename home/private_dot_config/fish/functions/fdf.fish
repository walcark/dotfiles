function fdf --description "fd + fzf — fuzzy pick a file, outputs path to stdout"
    fd $argv | fzf
end
