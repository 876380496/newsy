class Newsy < Formula
  desc "Terminal news reader with RSS and plugin-based sources"
  homepage "https://github.com/876380496/newsy"
  url "https://github.com/876380496/newsy/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "bd0b83fc87f96c6eacff8509372c75d5b1074d7c994acce71ffef6c7068516d1"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"newsy", "./cmd/newsy"
    pkgshare.install "config.yaml"
    (pkgshare/"plugins").install "plugins/echo-demo"
  end

  test do
    output = shell_output("#{bin}/newsy --print-paths")
    assert_match "config=", output
    assert_match "plugins=", output
    assert_match "db=", output
    assert_match "log=", output
  end
end
