import re
import os
import shutil

if __name__== "__main__":
    dir_path = os.path.dirname(os.path.realpath(__file__))
    seed_spec = os.path.join(dir_path, "seed.adoc")
    gen_seed_spec = os.path.join(dir_path, "seed.man.adoc")
    section_dir = os.path.join(dir_path, "sections")

    # process seed.adoc
    with open(gen_seed_spec, "w+") as manfile:
        with open(seed_spec, "r") as infile:
            intable = False
            columns = 0
            for line in infile:
                if re.match("= Seed Standard Definition", line):
                    line = line.replace("\n", "(1)\n")
                    manfile.write(line)
                # insert manpage attributes and Name section
                elif re.match(":docinfo:\n", line):
                    manfile.write(line)
                    manfile.write(":doctype: manpage\n")
                    manfile.write(":manmanual: Seed Specification\n")
                    manfile.write(":mansource: Seed Specification\n\n")
                    manfile.write("== Name\n\n")
                    manfile.write("seed-spec - A general standard to aid in the discovery and consumption of a discrete unit of work contained within a Docker image.\n")
                
                # Replace include with to be created manpage includes
                # elif re.match("include::sections", line):
                #     idx = line.index(".adoc")
                #     line = line[:idx] + ".man" + line[idx:]
                #     line = line.replace("sections", "sections-man")
                #     manfile.write(line)

                # find the start of a table    
                elif re.match("\[cols", line) and re.match("\|=+", next(infile, '')):
                    #no-op
                    intable = True
                    columns = len(re.findall("[\d*,]+\d", line)[0].split(","))

                    #skip the header line
                    next(infile, '')
                    next(infile, '')

                # find the bottom of the table
                elif intable and re.match("\|=+", line):
                    #no-op
                    intable = False

                # Write the lines of the table
                elif intable and re.match("\|\w*", line):
                    for c in range(columns):
                        if c == 0:
                            line = line.replace("|","").replace("\n","")
                            manfile.write("*"+line+"* ::\n")
                            line = next(infile)
                        else:
                            manfile.write("\t"+line.replace("|",""))
                            line = next(infile)
                
                # just write the line
                else:
                    manfile.write(line)

    # move generated seed.man.adoc to seed.adoc
    shutil.move(gen_seed_spec, seed_spec)

    for section_file_name in os.listdir(section_dir):
        if section_file_name.endswith(".adoc"):
            filename = "man." + section_file_name
            with open(os.path.join(section_dir, filename), "w+") as outfile:
                with open(os.path.join(section_dir, section_file_name)) as infile:
                    intable = False
                    columns = 0
                    inrow = False
                    for line in infile:
                        # ignore 
                        if re.match(":tabletags", line) or re.match("//", line):
                            # no-op
                            continue
                        # find the start of a table
                        if re.match("\[cols", line) and re.match("\|=+", next(infile)):
                            intable = True
                            columns = len(re.findall("[\d*,]+\d", line)[0].split(","))

                            #skip the header line
                            next(infile, '')
                            # next(infile, '')

                        # find the end of the table
                        elif intable and re.match("\|=+", line):
                            intable = False
                            inrow = False

                        # process the table
                        # term
                        elif intable and re.match("\|`\w*`", line):
                            inrow = True
                            line = line.replace("|","").replace("`","*").replace("\n","")
                            outfile.write("\n"+line+" ::\n")

                        # required
                        elif intable and inrow and re.match("\|w*", line):
                            line = line.replace("|","").replace("\n",". ")
                            outfile.write(line)
                        
                        # column width + Defintion
                        elif intable and inrow and re.match("\d\+[\w]*\|w*", line):
                            # strip the d+| off
                            substr = re.findall("\d\+[\w]*\|", line)[0]
                            line = line.replace(substr, "").replace("\n"," ")
                            outfile.write(line)

                        elif intable and inrow and line and re.match("\w+",line):
                            outfile.write(line)

                        # found the end of the 'row'
                        elif intable and inrow and (not line or line=="\n"):
                            inrow = False
                            outfile.write("\n")

                        elif intable and not inrow and re.match("[\d]*[\+]*\w\|", line):
                            if re.match("[\d]*[\+]*\w\|\s*_The following annotated snippet", line):
                                substr = re.findall("\d\+[\w]*\|\s*", line)[0]
                                line = line.replace(substr, "\n")
                                outfile.write(line)
                            else:
                                outfile.write("\n")
                            
                        elif intable and inrow:
                            outfile.write(line)
                        else:
                            outfile.write(line)
                # move generated file to overwrite original
                shutil.move(os.path.join(section_dir, filename), os.path.join(section_dir, section_file_name))

