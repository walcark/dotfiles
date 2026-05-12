function lspath --description "Manage colon-separated path variables"
    set var PATH
    set opt $argv[1]

    switch $opt
        case -h --help
            echo "lspath — manage path variables"
            echo ""
            echo "Usage: lspath [VAR]            list entries"
            echo "       lspath -a <entry> [VAR]  add after"
            echo "       lspath -b <entry> [VAR]  add before"
            echo "       lspath -r <entry> [VAR]  remove all occurrences"
            echo "       lspath -c [VAR]           remove non-existent + duplicates"

        case -a
            set entry $argv[2]
            test (count $argv) -ge 3 && set var $argv[3]
            if not contains $entry $$var
                set -gx $var $$var $entry
            end
            lspath $var

        case -b
            set entry $argv[2]
            test (count $argv) -ge 3 && set var $argv[3]
            set -gx $var $entry (string match -v $entry $$var)
            lspath $var

        case -r
            set entry $argv[2]
            test (count $argv) -ge 3 && set var $argv[3]
            set -gx $var (string match -v $entry $$var)
            lspath $var

        case -c
            test (count $argv) -ge 2 && set var $argv[2]
            set cleaned
            for p in $$var
                test -z $p && continue
                test ! -d $p && continue
                contains $p $cleaned && continue
                set cleaned $cleaned $p
            end
            set -gx $var $cleaned
            lspath $var

        case ''
            set var PATH

        case '*'
            # Treat first arg as variable name
            set var $opt
    end

    # Pretty print
    if test $opt != -a -a $opt != -b -a $opt != -r -a $opt != -c
        set green \e\[32m
        set red \e\[31m
        set yellow \e\[33m
        set bold \e\[1m
        set reset \e\[0m

        set entries $$var
        set total (count $entries)
        printf "$bold%s$reset (%d entries)\n" $var $total

        set i 1
        for entry in $entries
            if test -d $entry
                printf "  %2d  $green%s$reset\n" $i $entry
            else
                printf "  %2d  $red%s  ✗$reset\n" $i $entry
            end
            set i (math $i + 1)
        end
    end
end
