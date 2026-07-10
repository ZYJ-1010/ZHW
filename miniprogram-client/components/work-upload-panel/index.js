Component({
  options: {
    addGlobalClass: true
  },

  properties: {
    panelClass: {
      type: String,
      value: ''
    },
    imageText: {
      type: String,
      value: '上传图片'
    },
    fileText: {
      type: String,
      value: '上传文件'
    }
  },

  methods: {
    handleImageTap() {
      this.triggerEvent('imageupload')
    },

    handleFileTap() {
      this.triggerEvent('fileupload')
    }
  }
})
