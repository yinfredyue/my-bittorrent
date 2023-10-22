import subprocess

TEST_SCRIPT = "./your_bittorrent.sh"


class TestCase:
    def __init__(self, inputs, expected):
        self.inputs = inputs
        self.expected = expected


class TestSet:
    def __init__(self, description, testcases):
        self.description = description
        self.testcases = testcases


tests = {
    1: TestSet("Decode bencoded strings", [
        TestCase("decode 5:mango", '"mango"')
    ]),
    2: TestSet("Decode bencoded integers", [
        TestCase("decode i69640597e", '69640597'),
        TestCase("decode i42949678300e", '42949678300')
    ]),
    3: TestSet("Decode bencoded lists", [
        TestCase("decode l5:applei472ee", '["apple",472]'),
        TestCase("decode lli472e5:appleee", '[[472,"apple"]]')
    ]),
    4: TestSet("Decode bencoded dictionaries", [
        TestCase("decode de", '{}'),
        TestCase("decode d3:foo10:strawberry5:helloi52ee", '{"foo":"strawberry","hello":52}'),
        TestCase("decode d10:inner_dictd4:key16:value14:key2i42e8:list_keyl5:item15:item2i3eeee",
                 '{"inner_dict":{"key1":"value1","key2":42,"list_key":["item1","item2",3]}}')
    ])
}


for stage, testset in tests.items():
    print(f"[stage-{stage}] Running tests for Stage #2: {testset.description}")
    for testcase in testset.testcases:
        print(f"[stage-{stage}] Running {TEST_SCRIPT} {testcase.inputs}")
        print(f"[stage-{stage}] Expected output: {testcase.expected}")
        args = testcase.inputs.split()
        result = subprocess.run([TEST_SCRIPT] + args, stdout=subprocess.PIPE)
        output = result.stdout.decode('utf-8').strip()
        assert (result.returncode == 0)
        print(f"[your_program] {output}")
        assert (output == testcase.expected)

    print(f"[stage-{stage}] Test passed.\n")
