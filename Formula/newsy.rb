class Newsy < Formula
  desc "Terminal news reader with RSS and plugin-based sources"
  homepage "https://github.com/876380496/newsy"
  url "https://github.com/876380496/newsy/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_REAL_SHA256"
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
