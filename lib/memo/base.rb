module Memo
  class Base
    def initialize
      home        = ENV['HOME']
      memo_dir    = File.join(home, '.memo')
      @memo_file  = File.join(home, '.memo/memos')

      unless File.exists?(@memo_file)
        unless File.exists?(memo_dir)
          Dir.mkdir memo_dir
          File.new(@memo_file, 'w+')
        end
      end
    end

    def push(arg)
      File.open(@memo_file,'a') do |f|
        f.puts arg
      end
    end

    def pop
      File.open(@memo_file,'r') do |f|
        f.readlines.reverse.each do |line|
          print line
          STDIN.readline
        end
      end
    end
  end
end