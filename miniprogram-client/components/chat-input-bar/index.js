const EMOJI_LIST = ['😀', '😊', '😂', '😍', '😎', '😭', '😡', '👍', '👏', '🙏', '🎉', '❤️', '🔥', '✨', '😅', '🤝']

Component({
  properties: {
    placeholder: {
      type: String,
      value: ''
    }
  },

  data: {
    inputValue: '',
    isRecording: false,
    activePanel: '',
    emojis: EMOJI_LIST
  },

  lifetimes: {
    attached() {
      this.initRecorder()
    },

    detached() {
      if (this.data.isRecording) {
        this.stopRecord()
      }
    }
  },

  methods: {
    handleToolTap(event) {
      const { type } = event.currentTarget.dataset

      this.triggerEvent('tooltap', { type })
    },

    initRecorder() {
      if (!wx.getRecorderManager) {
        return
      }

      this.recorder = wx.getRecorderManager()

      this.recorder.onStart(() => {
        this.setData({
          isRecording: true,
          activePanel: ''
        })
        this.triggerEvent('recordstart')
      })

      this.recorder.onStop((result) => {
        this.setData({
          isRecording: false
        })
        this.triggerEvent('recordstop', result)
      })

      this.recorder.onError((error) => {
        this.setData({
          isRecording: false
        })
        this.triggerEvent('recorderror', error)
        wx.showToast({
          title: '录音失败',
          icon: 'none'
        })
      })
    },

    handleVoiceTap() {
      this.handleToolTap({
        currentTarget: {
          dataset: {
            type: 'voice'
          }
        }
      })

      if (this.data.isRecording) {
        this.stopRecord()
        return
      }

      this.startRecord()
    },

    startRecord() {
      if (!this.recorder) {
        wx.showToast({
          title: '当前环境不支持录音',
          icon: 'none'
        })
        return
      }

      wx.authorize({
        scope: 'scope.record',
        success: () => {
          this.recorder.start({
            duration: 60000,
            format: 'mp3'
          })
        },
        fail: () => {
          wx.showModal({
            title: '需要录音权限',
            content: '请在设置中开启麦克风权限后再录音',
            confirmText: '去设置',
            success: (result) => {
              if (result.confirm && wx.openSetting) {
                wx.openSetting()
              }
            }
          })
        }
      })
    },

    stopRecord() {
      if (this.recorder) {
        this.recorder.stop()
      }
    },

    handleInput(event) {
      const value = event.detail.value || ''

      this.setData({
        inputValue: value
      })
      this.triggerEvent('inputchange', { value })
    },

    handleInputFocus() {
      this.setData({
        activePanel: ''
      })
      this.triggerEvent('inputfocus')
    },

    handleInputConfirm() {
      this.sendText()
    },

    sendText() {
      const value = String(this.data.inputValue || '').trim()

      if (!value) {
        return
      }

      this.triggerEvent('sendtext', {
        value
      })

      this.setData({
        inputValue: '',
        activePanel: ''
      })
    },

    handleEmojiTap() {
      this.setData({
        activePanel: this.data.activePanel === 'emoji' ? '' : 'emoji'
      })
      this.triggerEvent('panelchange', {
        panel: this.data.activePanel
      })
      this.triggerEvent('tooltap', { type: 'emoji' })
    },

    handleEmojiSelect(event) {
      const { emoji } = event.currentTarget.dataset
      const inputValue = `${this.data.inputValue}${emoji}`

      this.setData({
        inputValue
      })
      this.triggerEvent('emojiselect', { emoji, value: inputValue })
      this.triggerEvent('inputchange', { value: inputValue })
    },

    handleSendOrAttachTap() {
      if (String(this.data.inputValue || '').trim()) {
        this.sendText()
        return
      }

      this.setData({
        activePanel: this.data.activePanel === 'attach' ? '' : 'attach'
      })
      this.triggerEvent('panelchange', {
        panel: this.data.activePanel
      })
      this.triggerEvent('tooltap', { type: 'add' })
    },

    handleChooseImage() {
      if (wx.chooseMedia) {
        const result = wx.chooseMedia({
          count: 9,
          mediaType: ['image'],
          sourceType: ['album', 'camera'],
          success: (response) => {
            this.triggerEvent('chooseimage', response)
          }
        })

        if (result && result.then) {
          result.then((response) => {
            this.triggerEvent('chooseimage', response)
          })
        }
        return
      }

      wx.chooseImage({
        count: 9,
        sourceType: ['album', 'camera'],
        success: (result) => {
          this.triggerEvent('chooseimage', result)
        }
      })
    },

    handleChooseFile() {
      if (!wx.chooseMessageFile) {
        wx.showToast({
          title: '当前环境不支持选择文件',
          icon: 'none'
        })
        return
      }

      wx.chooseMessageFile({
        count: 5,
        type: 'file',
        extension: ['pdf', 'txt', 'zip'],
        success: (result) => {
          this.triggerEvent('choosefile', result)
        }
      })
    }
  }
})
