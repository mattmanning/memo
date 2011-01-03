module Memo
  class Base
    def initialize
      home        = ENV['HOME']
      memo_dir    = File.join(home, '.memo')
      @memo_file  = File.join(home, '.memo/memos')
      @memo_stack = File.open(@memo_file, 'r').readlines

      create_memo_file!
    end

    def push(arg)
      File.open(@memo_file,'a') do |f|
        f.puts arg
      end
    end

    def pop
      @memo_stack.pop
    end

    private

    def create_memo_file!
      unless File.exists?(@memo_file)
        unless File.exists?(memo_dir)
          Dir.mkdir memo_dir
          File.new(@memo_file, 'w+')
        end
      end
    end
  end
end