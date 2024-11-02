import subprocess
import os

goos_options = ["linux"]
goarch_options = ["arm64"]
# goarch_options = ["arm64", "mips", "mipsle"]
# gomips_options = ["hardfloat", "softfloat"]
gomips_options = [None]
output_binary = "ip_changer"

def compile_program(goos, goarch, gomips=None):
    os.environ["GOOS"] = goos
    os.environ["GOARCH"] = goarch
    if gomips:
        os.environ["GOMIPS"] = gomips
    else:
        os.environ.pop("GOMIPS", None)

    command = f"go build -o {output_binary}_{goarch}_{gomips if gomips else 'default'} main.go"

    try:
        print(f"Compiling with GOOS={goos}, GOARCH={goarch}, GOMIPS={gomips if gomips else 'N/A'}")
        result = subprocess.run(command, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        print(f"Success: {result.stdout.decode('utf-8')}")
    except subprocess.CalledProcessError as e:
        print(f"Failed: {e.stderr.decode('utf-8')}")

def main():
    for goos in goos_options:
        for goarch in goarch_options:
            if goarch == "mips" or goarch == "mipsle":
                for gomips in gomips_options:
                    compile_program(goos, goarch, gomips)
            else:
                compile_program(goos, goarch)

if __name__ == "__main__":
    main()
