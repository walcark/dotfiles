function fdo --description "fd + fzf + open — pick a file and open it"
    set open_cmd xdg-open
    if test "$argv[1]" = "-o"
        set open_cmd $argv[2]
        set argv $argv[3..-1]
    end
    set result (fd $argv | fzf)
    test -n "$result" && $open_cmd $result
end
