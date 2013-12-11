package cli

import (
  "dogestry/client"
  "dogestry/remote"
  "encoding/json"
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
)

func (cli *DogestryCli) CmdPull(args ...string) error {
  cmd := cli.Subcmd("push", "REMOTE IMAGE[:TAG]", "pull IMAGE from the REMOTE and load it into docker. TAG defaults to 'latest'")
  if err := cmd.Parse(args); err != nil {
    return nil
  }

  if len(cmd.Args()) < 2 {
    return fmt.Errorf("Error: REMOTE and IMAGE not specified")
  }

  remoteDef := cmd.Arg(0)
  image := cmd.Arg(1)

  imageRoot, err := cli.WorkDir(image)
  if err != nil {
    return err
  }
  r, err := remote.NewRemote(remoteDef, cli.Config)
  if err != nil {
    return err
  }

  fmt.Println("remote", r.Desc())

  fmt.Println("resolving image id")
  id, err := r.ResolveImageNameToId(image)
  if err != nil {
    return err
  }

  fmt.Printf("image '%s' resolved on remote id '%s'\n", image, id.Short())

  fmt.Println("preparing images")
  if err := cli.preparePullImage(id, imageRoot, r); err != nil {
    return err
  }

  fmt.Println("preparing repositories file")
  if err := prepareRepositories(image, imageRoot, r); err != nil {
    return err
  }

  fmt.Println("sending tar to docker")
  if err := cli.sendTar(imageRoot); err != nil {
    return err
  }

  return nil
}

func (cli *DogestryCli) preparePullImage(fromId remote.ID, imageRoot string, r remote.Remote) error {
  return r.WalkImages(fromId, func(id remote.ID, image client.Image, err error) error {
    fmt.Printf("examining id '%s' on remote\n", id.Short())
    if err != nil {
      fmt.Println("err", err)
      return err
    }

    _, err = cli.client.InspectImage(string(id))
    if err == client.ErrNoSuchImage {
      return cli.pullImage(id, filepath.Join(imageRoot, string(id)), r)
    } else {
      fmt.Printf("docker already has id '%s', stopping\n", id.Short())
      return remote.BreakWalk
    }
  })
}

func (cli *DogestryCli) pullImage(id remote.ID, dst string, r remote.Remote) error {
  err := r.PullImageId(id, dst)
  if err != nil {
    return err
  }
  return cli.processPulled(id, dst)
}

func (cli *DogestryCli) processPulled(id remote.ID, dst string) error {
  compressedLayerFile := filepath.Join(dst, "layer.tar.lz4")
  return cli.compressor.Decompress(compressedLayerFile)
}

func prepareRepositories(image, imageRoot string, r remote.Remote) error {
  repoName, repoTag := remote.NormaliseImageName(image)

  id, err := r.ParseTag(repoName, repoTag)
  if err != nil {
    return err
  } else if id == "" {
    return nil
  }

  reposPath := filepath.Join(imageRoot, "repositories")
  reposFile, err := os.Create(reposPath)
  if err != nil {
    return err
  }
  defer reposFile.Close()

  repositories := map[string]Repository{}
  repositories[repoName] = Repository{}
  repositories[repoName][repoTag] = string(id)

  return json.NewEncoder(reposFile).Encode(&repositories)
}

// stream the tarball into docker
// its easier here to use tar command, but it'd be neater to mirror Push's approach
func (cli *DogestryCli) sendTar(imageRoot string) error {
  notExist,err := dirNotExistOrEmpty(imageRoot)

  if err != nil {
    return err
  }
  if notExist {
    fmt.Println("no images to send to docker")
    return nil
  }


  cmd := exec.Command("/bin/tar", "cvf", "-", ".")
  cmd.Dir = imageRoot
  defer cmd.Wait()

  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return err
  }

  if err := cmd.Start(); err != nil {
    return err
  }

  fmt.Println("kicking off post")
  return cli.client.PostImageTarball(stdout)
}

func dirNotExistOrEmpty(path string) (bool,error) {
  imagesDir, err := os.Open(path)
  if err != nil {
    // no images
    if os.IsNotExist(err) {
      return true,nil
    } else {
      return false,err
    }
  }
  defer imagesDir.Close()

  names, err := imagesDir.Readdirnames(-1)
  if err != nil {
    return false,err
  }


  if len(names) <= 0 {
    return true,nil
  }

  return false, nil
}
